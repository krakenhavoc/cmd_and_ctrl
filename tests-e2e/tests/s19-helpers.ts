// S19 e2e helper layer. Three responsibilities:
//
//   1. Drive the lobby HTTP API to spin up a 2-player game seeded
//      with the S19 caster + opponent decks.
//   2. Maintain a Node-side admin WebSocket connection that lets the
//      tests admin-move cards across zones (skipping mana / phase
//      gates) and observe authoritative snapshots — the test sees
//      the same `pending_choices` array every player browser sees.
//   3. Provide name-keyed lookups so the test layer can talk in
//      "find me an opponent's Sol Ring" rather than instance UUIDs.
//
// The helper avoids mocking: every assertion is on a real WS
// snapshot pushed by the server after a real action passed
// through the same Dispatch path the production client uses.

import type { APIRequestContext, Browser, BrowserContext, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import {
  adminLogin,
  createGame,
  uploadDeckAs,
  startGameAs,
  type GameMeta,
} from "./lobby-api";
import { SERVER_ORIGIN, SERVER_WS_ORIGIN } from "./env";
import { makeS19CasterDeck, makeS19OpponentDeck } from "./s19-deck-fixture";

// --- Snapshot view types --------------------------------------
//
// Loose mirrors of server/internal/protocol/view.go shapes. Only
// fields the tests actually read are typed; everything else flows
// through as `unknown`. Strict typing here would force a churn pass
// on every additive wire change in unrelated sprints.

export interface SnapshotCard {
  instance_id: string;
  name?: string;
  type_line?: string;
  controller?: string;
  owner?: string;
  tapped?: boolean;
  // Player ID this card is declared to attack. #328's spec reads it
  // to confirm an attack landed before the blocking window opens.
  attacking_target?: string;
  // #2219: the non-keyword abilities behind an art tile's chips.
  ability_rows?: { kind: string; label: string }[];
}

export interface SnapshotZone {
  kind: string;
  owner?: string;
  count: number;
  cards: SnapshotCard[];
}

export interface SnapshotPlayer {
  id: string;
  name: string;
  seat: number;
  hand: SnapshotZone;
  library: SnapshotZone;
  graveyard: SnapshotZone;
  command: SnapshotZone;
  // CR 103.5 (#2237): the one seat whose turn it is to keep or mulligan.
  mulligan_turn?: boolean;
}

export interface PendingChoice {
  id: string;
  kind: string;
  chooser: string;
  source?: string;
  reason?: string;
  pay_cost?: string;
}

export interface SnapshotStackItem {
  id: string;
  kind: string;
  controller: string;
  source_card_id: string;
  label?: string;
}

export interface SnapshotTurn {
  active_seat: number;
  priority_holder: number;
  step: string;
  // #328: seat indices that owe a declare-blockers decision. Present
  // only during declare_blockers.
  block_decision_seats?: number[];
  // #1279: defending seats whose block declaration is finished (with
  // or without blocks). Present only during declare_blockers.
  blocks_declared_seats?: number[];
  // #1279: defending seats still declaring. While any is, #1501 parks
  // priority (priority_holder -1). Present only during declare_blockers.
  block_pending_seats?: number[];
}

export interface SnapshotView {
  state: string;
  mulligans_open?: boolean;
  seats: SnapshotPlayer[];
  battlefield: SnapshotZone;
  exile: SnapshotZone;
  turn?: SnapshotTurn;
  // Spells AND ability items on the stack, bottom..top. S19
  // triggers show up here (kind "triggered") with no card in the
  // stack zone — see viewOfStackItemsInStackOrder server-side.
  stack_items?: SnapshotStackItem[];
  pending_choices?: PendingChoice[];
}

// --- Admin WebSocket client ----------------------------------
//
// The hub wholesale-hides hand and library contents for an admin
// connected without a player binding (admin spectator), so the test
// layer can't find a seat's library cards by name from a single
// admin WS. We side-step that by opening one admin-token WS bound to
// each seat — admin-token + ?player=<id> grants admin's
// no-priority-required dispatch with the bound seat's full
// visibility (CR 400.2 / hand-and-library private-zone reveal). The
// two bound views are joined into one merged snapshot for lookups.

export interface AdminClient {
  // sendActionAsPlayer emits an action frame routed through the
  // admin connection bound to `playerID`. The frame carries
  // `player: <playerID>` in its payload so per-seat actions
  // (keep_hand / pass_priority / move_card affecting that seat's
  // private zones) authenticate. The admin role bypasses the
  // priority / phase gates a seated session would hit.
  sendActionAsPlayer(
    playerID: string,
    type: string,
    params: Record<string, unknown>,
  ): Promise<SnapshotView>;

  // sendAction emits an action frame WITHOUT a player binding —
  // useful for moves where the seat doesn't matter to the gate
  // (admin direct move_card across any zones). Routes through the
  // caster-bound admin WS by convention.
  sendAction(type: string, params: Record<string, unknown>): Promise<SnapshotView>;

  // snapshot returns a merged view: shared zones (battlefield,
  // exile, stack) come from the caster's WS; each seat's private
  // zones (hand, library, graveyard, command) come from the WS
  // bound to that seat. Throws when called before initial
  // snapshots from both connections have arrived.
  snapshot(): SnapshotView;

  // waitFor polls the merged snapshot until `predicate` returns
  // true, using either connection's events to advance.
  waitFor(
    predicate: (v: SnapshotView) => boolean,
    message: string,
    timeoutMs?: number,
  ): Promise<SnapshotView>;

  // cardName returns the name of a card instance as the server itself
  // recorded it, or undefined when the server has no such card. Unlike
  // snapshot(), it can name a card in a library, which every view on
  // the wire redacts (CR 400.2). See cardNamesFromReplay.
  cardName(instanceID: string): Promise<string | undefined>;

  close(): void;
}

// AdminConnection is one of the two underlying WS connections. Not
// part of the public API; consumed by openAdminClient to assemble
// the merged AdminClient.
interface AdminConnection {
  send(type: string, params: Record<string, unknown>, asPlayer?: string): Promise<SnapshotView>;
  snapshot(): SnapshotView;
  onSnapshot(cb: (v: SnapshotView) => void): () => void;
  close(): void;
}

interface AdminWSFrame {
  kind: string;
  id?: string;
  payload?: { game?: SnapshotView; code?: string; message?: string };
}

async function openAdminConnection(
  adminToken: string,
  gameID: string,
  asSeatID: string,
): Promise<AdminConnection> {
  const url = `${SERVER_WS_ORIGIN}/ws?game=${gameID}&player=${encodeURIComponent(asSeatID)}&token=${encodeURIComponent(adminToken)}`;
  // Node-side WebSocket: global since Node 22. Node 20 throws a bare
  // ReferenceError here, which is how the nightly went red in Sept 2026.
  if (typeof WebSocket === "undefined") {
    throw new Error(
      `global WebSocket is unavailable (Node ${process.version}); tests-e2e requires Node >= 22`,
    );
  }
  const ws = new WebSocket(url);
  let latest: SnapshotView | null = null;
  type Listener = (v: SnapshotView) => void;
  const subscribers: Listener[] = [];
  type ErrorListener = (err: { id?: string; code: string; message: string }) => void;
  const errorSubscribers: ErrorListener[] = [];

  await new Promise<void>((resolve, reject) => {
    const t = setTimeout(() => reject(new Error("admin WS open timed out")), 8000);
    ws.addEventListener(
      "open",
      () => {
        clearTimeout(t);
        resolve();
      },
      { once: true },
    );
    ws.addEventListener(
      "error",
      () => {
        clearTimeout(t);
        reject(new Error("admin WS open error"));
      },
      { once: true },
    );
  });

  ws.addEventListener("message", (ev) => {
    let frame: AdminWSFrame;
    try {
      frame = JSON.parse(String(ev.data)) as AdminWSFrame;
    } catch {
      return;
    }
    if (frame.kind === "snapshot" && frame.payload?.game) {
      latest = frame.payload.game;
      for (const cb of subscribers.slice()) cb(latest);
    } else if (frame.kind === "error" && frame.payload) {
      for (const cb of errorSubscribers.slice())
        cb({
          id: frame.id,
          code: frame.payload.code ?? "unknown",
          message: frame.payload.message ?? "",
        });
    }
  });

  await new Promise<void>((resolve, reject) => {
    if (latest) {
      resolve();
      return;
    }
    const t = setTimeout(() => reject(new Error("initial snapshot timed out")), 8000);
    subscribers.push(function once() {
      clearTimeout(t);
      const idx = subscribers.indexOf(once);
      if (idx >= 0) subscribers.splice(idx, 1);
      resolve();
    });
  });

  function nextSnapshot(timeoutMs: number, frameID: string): Promise<SnapshotView> {
    return new Promise((resolve, reject) => {
      const t = setTimeout(() => {
        cleanup();
        reject(new Error(`waiting for snapshot timed out (id=${frameID})`));
      }, timeoutMs);
      const onSnap = (v: SnapshotView) => {
        cleanup();
        resolve(v);
      };
      const onErr = (err: { id?: string; code: string; message: string }) => {
        if (err.id && err.id !== frameID) return;
        cleanup();
        reject(new Error(`server rejected ${frameID}: ${err.code}: ${err.message}`));
      };
      const cleanup = () => {
        clearTimeout(t);
        const i1 = subscribers.indexOf(onSnap);
        if (i1 >= 0) subscribers.splice(i1, 1);
        const i2 = errorSubscribers.indexOf(onErr);
        if (i2 >= 0) errorSubscribers.splice(i2, 1);
      };
      subscribers.push(onSnap);
      errorSubscribers.push(onErr);
    });
  }

  return {
    async send(type, params, asPlayer) {
      const id = crypto.randomUUID();
      const payload: Record<string, unknown> = { type, params };
      if (asPlayer) payload.player = asPlayer;
      const frame = { v: 0, kind: "action", id, payload };
      const promise = nextSnapshot(8000, id);
      ws.send(JSON.stringify(frame));
      return await promise;
    },
    snapshot() {
      if (!latest) throw new Error("no snapshot received yet");
      return latest;
    },
    onSnapshot(cb) {
      subscribers.push(cb);
      return () => {
        const idx = subscribers.indexOf(cb);
        if (idx >= 0) subscribers.splice(idx, 1);
      };
    },
    close() {
      ws.close();
    },
  };
}

// cardNamesFromReplay maps every card instance in the game to its
// name, read from the server's own record of the game.
//
// Why: a library is hidden on the wire (CR 400.2). The seat-bound
// admin view keeps each library card's instance_id but strips its
// name, so the wire alone cannot say which library card is the Sol
// Ring. The helpers used to find out by drawing until the card
// surfaced. That flooded the hand with up to ninety cards and sent the
// browsers a snapshot per draw, and on a shared runner the players'
// pages fell 10-20 seconds behind the server re-rendering them. Every
// UI assertion made after the seeding then raced that backlog (#2253).
//
// The admin's replay download (GET /games/{id}/replay) is the
// server's unfiltered record: one SnapshotPayload per committed
// action, every card named. Instance IDs never change when a card
// changes zone, so one read gives a map that stays true for the rest
// of the game. The replay is appended under the room lock before the
// snapshot is broadcast, so a card the admin has seen in a snapshot is
// already in the file.
async function cardNamesFromReplay(
  adminToken: string,
  gameID: string,
): Promise<Map<string, string>> {
  const res = await fetch(`${SERVER_ORIGIN}/games/${gameID}/replay`, {
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  if (!res.ok) {
    throw new Error(`replay download failed: ${res.status} ${await res.text()}`);
  }
  const lines = (await res.text()).split("\n").filter((l) => l.trim() !== "");
  const last = lines.at(-1);
  if (!last) throw new Error(`replay for ${gameID} is empty`);
  const game = (JSON.parse(last) as { game: SnapshotView }).game;
  const names = new Map<string, string>();
  const add = (z: SnapshotZone | undefined) => {
    for (const c of z?.cards ?? []) {
      if (c.name) names.set(c.instance_id, c.name);
    }
  };
  for (const seat of game.seats) {
    add(seat.library);
    add(seat.hand);
    add(seat.graveyard);
    add(seat.command);
  }
  add(game.battlefield);
  add(game.exile);
  return names;
}

// openAdminClient opens two admin-token WS connections — one bound
// to each seat — so the test layer has full visibility into both
// players' private zones (hand / library). Snapshot reads are
// merged: shared zones (battlefield, exile, stack, pending_choices)
// come from the caster connection by convention; each seat's
// private zones are pulled from the WS bound to that seat.
export async function openAdminClient(
  adminToken: string,
  gameID: string,
  casterID: string,
  opponentID: string,
): Promise<AdminClient> {
  const [casterConn, opponentConn] = await Promise.all([
    openAdminConnection(adminToken, gameID, casterID),
    openAdminConnection(adminToken, gameID, opponentID),
  ]);

  function mergedSnapshot(): SnapshotView {
    const cv = casterConn.snapshot();
    const ov = opponentConn.snapshot();
    const seats: SnapshotPlayer[] = cv.seats.map((seat) => {
      // Use the seat's bound view for its own private zones — that's
      // the only WS that sees them un-redacted.
      if (seat.id === casterID) {
        return seat;
      }
      const fromOpp = ov.seats.find((s) => s.id === seat.id);
      return fromOpp ? fromOpp : seat;
    });
    return { ...cv, seats };
  }

  type Listener = (v: SnapshotView) => void;
  const subscribers: Listener[] = [];
  const broadcast = (v: SnapshotView) => {
    for (const cb of subscribers.slice()) cb(v);
  };
  casterConn.onSnapshot(() => broadcast(mergedSnapshot()));
  opponentConn.onSnapshot(() => broadcast(mergedSnapshot()));

  // Read once, on first use: a card's name and instance ID never
  // change, so the map does not go stale. A miss re-reads it, which
  // covers a card created after the first read (a token, a spawn).
  let names: Map<string, string> | null = null;

  return {
    async sendAction(type, params) {
      // Route through the caster connection. Either would work — we
      // just need a connected admin WS to dispatch.
      const v = await casterConn.send(type, params);
      // Return the MERGED snapshot, not the per-connection one. The
      // sendAction-from-caster snapshot omits the opponent's private
      // zones; the merged one is what callers expect.
      return mergedSnapshot();
    },
    async sendActionAsPlayer(playerID, type, params) {
      const conn = playerID === casterID ? casterConn : opponentConn;
      await conn.send(type, params, playerID);
      return mergedSnapshot();
    },
    snapshot() {
      return mergedSnapshot();
    },
    async waitFor(predicate, message, timeoutMs = 5000) {
      const cur = mergedSnapshot();
      if (predicate(cur)) return cur;
      return await new Promise<SnapshotView>((resolve, reject) => {
        const t = setTimeout(() => {
          const idx = subscribers.indexOf(cb);
          if (idx >= 0) subscribers.splice(idx, 1);
          reject(new Error(`waitFor timed out: ${message}`));
        }, timeoutMs);
        const cb = (v: SnapshotView) => {
          if (predicate(v)) {
            clearTimeout(t);
            const idx = subscribers.indexOf(cb);
            if (idx >= 0) subscribers.splice(idx, 1);
            resolve(v);
          }
        };
        subscribers.push(cb);
      });
    },
    async cardName(instanceID) {
      if (!names?.has(instanceID)) {
        names = await cardNamesFromReplay(adminToken, gameID);
      }
      return names.get(instanceID);
    },
    close() {
      casterConn.close();
      opponentConn.close();
    },
  };
}

// --- Card / zone lookups -------------------------------------

export function findCardInZone(zone: SnapshotZone, name: string): SnapshotCard | null {
  for (const c of zone.cards) {
    if (c.name === name) return c;
  }
  return null;
}

export function requireCardInZone(
  zone: SnapshotZone,
  name: string,
  context: string,
): SnapshotCard {
  const c = findCardInZone(zone, name);
  if (!c) {
    throw new Error(`expected ${name} in ${context}, found ${zone.cards.length} cards`);
  }
  return c;
}

export function playerByID(v: SnapshotView, playerID: string): SnapshotPlayer {
  const seat = v.seats.find((s) => s.id === playerID);
  if (!seat) throw new Error(`no seat with id ${playerID}`);
  return seat;
}

export function findCardInPlayerLibrary(
  v: SnapshotView,
  playerID: string,
  name: string,
): SnapshotCard | null {
  return findCardInZone(playerByID(v, playerID).library, name);
}

export function findCardInPlayerHand(
  v: SnapshotView,
  playerID: string,
  name: string,
): SnapshotCard | null {
  return findCardInZone(playerByID(v, playerID).hand, name);
}

export function findCardOnBattlefield(
  v: SnapshotView,
  name: string,
): SnapshotCard | null {
  return findCardInZone(v.battlefield, name);
}

export function findCardInPlayerGraveyard(
  v: SnapshotView,
  playerID: string,
  name: string,
): SnapshotCard | null {
  return findCardInZone(playerByID(v, playerID).graveyard, name);
}

// --- Move-card admin convenience -----------------------------

type OwnZoneKind = "library" | "hand" | "graveyard";

function ownZone(v: SnapshotView, ownerID: string, kind: OwnZoneKind): SnapshotZone {
  const seat = playerByID(v, ownerID);
  return kind === "library" ? seat.library : kind === "hand" ? seat.hand : seat.graveyard;
}

function zoneHolds(zone: SnapshotZone, instanceID: string): boolean {
  return zone.cards.some((c) => c.instance_id === instanceID);
}

// findCardInLibrary returns a card named `name` from ownerID's library,
// or null. The owner's library reaches the seat-bound view with each
// card's instance_id and no name (CR 400.2), so the name comes from
// the server's record (AdminClient.cardName).
export async function findCardInLibrary(
  admin: AdminClient,
  ownerID: string,
  name: string,
): Promise<SnapshotCard | null> {
  for (const c of playerByID(admin.snapshot(), ownerID).library.cards) {
    if ((await admin.cardName(c.instance_id)) === name) return c;
  }
  return null;
}

// moveOwnCard moves one of ownerID's cards and waits until the server
// confirms it left `src`. Waiting on the source rather than the
// destination keeps it true for a card that moves on at once (a
// permanent whose ETB sends it elsewhere).
//
// It routes through the owner's admin connection so action.Caller
// matches the card's controller: the move_card gate
// (`requireCardController`) only bypasses when Caller=uuid.Nil, which
// a seat-bound admin WS isn't.
async function moveOwnCard(
  admin: AdminClient,
  ownerID: string,
  instanceID: string,
  src: OwnZoneKind,
  dst: { kind: string; owner?: string },
  what: string,
): Promise<SnapshotView> {
  await admin.sendActionAsPlayer(ownerID, "move_card", {
    src: { kind: src, owner: ownerID },
    dst,
    instance_id: instanceID,
  });
  return await admin.waitFor(
    (v) => !zoneHolds(ownZone(v, ownerID, src), instanceID),
    `${what} left ${ownerID.slice(0, 8)}'s ${src}`,
  );
}

// seedHandWithCard puts a card named `name` into ownerID's hand and
// returns it. A copy already in hand is returned as it is; otherwise
// one is moved there straight from the library, in one action. Used by
// tests that need to know the hand size right BEFORE a trigger fires
// (the `adminMoveByName` shorthand bundles seeding and moving; this
// primitive splits them apart so tests can snapshot in between).
//
// It used to draw until the card surfaced, because the wire does not
// name library cards. That moved up to ninety cards into the hand and
// left the players' pages seconds behind the server (#2253).
export async function seedHandWithCard(
  admin: AdminClient,
  ownerID: string,
  name: string,
): Promise<SnapshotCard> {
  const already = findCardInZone(playerByID(admin.snapshot(), ownerID).hand, name);
  if (already) return already;
  const card = await findCardInLibrary(admin, ownerID, name);
  if (!card) {
    const owner = playerByID(admin.snapshot(), ownerID);
    throw new Error(
      `seedHandWithCard: no ${name} in ${ownerID.slice(0, 8)}'s library ` +
        `(library_count=${owner.library.count} hand_count=${owner.hand.count} ` +
        `graveyard_count=${owner.graveyard.count})`,
    );
  }
  await moveOwnCard(
    admin,
    ownerID,
    card.instance_id,
    "library",
    { kind: "hand", owner: ownerID },
    name,
  );
  const seeded = await admin.waitFor(
    (v) => zoneHolds(playerByID(v, ownerID).hand, card.instance_id),
    `${name} in ${ownerID.slice(0, 8)}'s hand`,
  );
  const inHand = playerByID(seeded, ownerID).hand.cards.find(
    (c) => c.instance_id === card.instance_id,
  );
  if (!inHand) throw new Error(`seedHandWithCard: ${name} vanished from the hand`);
  return inHand;
}

// adminMoveByName moves a card identified by name to the target zone,
// from wherever it is among the owner's hand, graveyard and library
// (searched in that order), in one action. It returns once the server
// confirms the card left where it was.
//
// `_srcHint` is informational only — kept on the signature so test
// call sites still document where the card "lives" conceptually.
export async function adminMoveByName(
  admin: AdminClient,
  ownerID: string,
  name: string,
  _srcHint: "library" | "hand" | "graveyard",
  dstKind: "battlefield" | "graveyard" | "hand" | "exile",
): Promise<SnapshotView> {
  const owner = playerByID(admin.snapshot(), ownerID);
  let card: SnapshotCard | null = findCardInZone(owner.hand, name);
  let src: OwnZoneKind = "hand";
  if (!card) {
    card = findCardInZone(owner.graveyard, name);
    src = "graveyard";
  }
  if (!card) {
    card = await findCardInLibrary(admin, ownerID, name);
    src = "library";
  }
  if (!card) {
    throw new Error(
      `adminMoveByName: no ${name} in ${ownerID.slice(0, 8)}'s hand, graveyard or library`,
    );
  }

  const dst: { kind: string; owner?: string } = { kind: dstKind };
  if (dstKind === "graveyard" || dstKind === "hand") {
    dst.owner = ownerID;
  }
  return await moveOwnCard(admin, ownerID, card.instance_id, src, dst, name);
}

// returnToLibrary moves up to `n` cards named `name` from ownerID's
// hand back onto their library, and returns the last snapshot. The
// Solemn test uses it to be certain the deck's one Island is in the
// library, where the search can offer it, even when the opening hand
// drew it.
export async function returnToLibrary(
  admin: AdminClient,
  ownerID: string,
  name: string,
  n: number,
): Promise<SnapshotView> {
  let last = admin.snapshot();
  let moved = 0;
  for (let i = 0; i < n; i++) {
    const card = findCardInZone(playerByID(admin.snapshot(), ownerID).hand, name);
    if (!card) break;
    last = await moveOwnCard(
      admin,
      ownerID,
      card.instance_id,
      "hand",
      { kind: "library", owner: ownerID },
      name,
    );
    moved++;
  }
  if (moved === 0) {
    throw new Error(`returnToLibrary: no ${name} in ${ownerID}'s hand to put back`);
  }
  return last;
}

// --- Browser orchestration -----------------------------------

export interface JoinedPlayer {
  context: BrowserContext;
  page: Page;
  name: string;
  token: string;
  playerID: string;
}

// S19_GAMEPLAY is what every S19 seat's settings start from.
//
// passMode "careful" (ADR 0143 §2.1; until v24 this was
// alwaysStopOpponentStack) makes the client hold on an opponent's stack
// item. Smart autopass (#1308) passes for a seat with nothing to
// respond with, so on an idle table the opponent passes the caster's
// trigger the moment it lands and it resolves before any assertion can
// see it; before that change an opponent's stack item always held.
// These tests assert on the trigger while it waits and resolve it
// through resolveStack's own "next" clicks.
//
// stackHoldMs: 0 turns off ADR 0119 §2's stack hold. These seats hold
// by hand anyway and resolve with `next`, which is never delayed; the
// hold has its own spec (stack-hold-2204.spec.ts).
//
// Everything else stays on the defaults (auto-pass through upkeep/draw
// is what setup waits for).
export const S19_GAMEPLAY: Record<string, unknown> = {
  passMode: "careful",
  stackHoldMs: 0,
};

async function joinAsPlayer(
  browser: Browser,
  gameID: string,
  inviteToken: string,
  name: string,
  gameplay: Record<string, unknown> = S19_GAMEPLAY,
  viewport?: { width: number; height: number },
): Promise<JoinedPlayer> {
  const context = await browser.newContext(viewport ? { viewport } : {});
  // Written only when absent so a later navigation does not undo what
  // the app saved.
  await context.addInitScript((seed) => {
    try {
      if (localStorage.getItem("cmdctrl.settings.v1") === null) {
        localStorage.setItem("cmdctrl.settings.v1", JSON.stringify({ gameplay: seed }));
      }
    } catch {
      // storage unavailable: fall back to the defaults
    }
  }, gameplay);
  const page = await context.newPage();
  await page.goto(`/#/games/${gameID}/join?t=${encodeURIComponent(inviteToken)}`);
  await page.getByPlaceholder("your name").fill(name);
  await page.getByRole("button", { name: "join" }).click();
  // The post-join redirect lands on the lobby first (per Join.svelte:60).
  // Wait for the session to be stamped — once localStorage has it, the
  // join API call has succeeded and the route has navigated. Polling
  // localStorage is more reliable than waiting on the URL string,
  // which can race with hash-router state.
  //
  // 30s timeout because consecutive tests in the suite stress the
  // shared dev server enough that the join handler can stall briefly
  // — the per-test sessions are independent, but the http server's
  // rate limiter and the WS hub's room registration share state
  // across runs.
  await page.waitForFunction(
    () => {
      const raw = localStorage.getItem("cmdctrl.session");
      if (!raw) return false;
      const s = JSON.parse(raw);
      return s?.token && s?.playerID;
    },
    null,
    { timeout: 30_000 },
  );
  const session = await page.evaluate(() =>
    JSON.parse(localStorage.getItem("cmdctrl.session") ?? "null"),
  );
  if (!session?.token || !session?.playerID) {
    throw new Error(`${name}: session missing after join`);
  }
  // Navigate into the game route directly. The lobby's "open" button
  // does the same `navigate("#/games/{id}")` — bypassing it keeps the
  // helper independent of lobby UI churn.
  await page.goto(`/#/games/${gameID}`);
  await expect(page).toHaveURL(new RegExp(`#/games/${gameID}$`), { timeout: 10_000 });
  return { context, page, name, token: session.token, playerID: session.playerID };
}

export interface S19Setup {
  game: GameMeta;
  adminToken: string;
  admin: AdminClient;
  caster: JoinedPlayer;
  opponent: JoinedPlayer;
  // shutdown closes both browser contexts and the admin WS. Tests
  // call this in a final cleanup step (or via test.afterEach when
  // the suite uses a beforeAll fixture).
  shutdown(): Promise<void>;
}

// triggerOnStack returns the triggered-ability stack item sourced
// from the named battlefield card, or null. S19 triggers sit on the
// stack (kind "triggered", no card in the stack zone) until every
// player passes priority in succession — tests assert the trigger
// is waiting here before resolving it with resolveStack.
export function triggerOnStack(v: SnapshotView, sourceName: string): SnapshotStackItem | null {
  const src = findCardOnBattlefield(v, sourceName) ?? findCardAnywhere(v, sourceName);
  if (!src) return null;
  return (
    (v.stack_items ?? []).find(
      (it) => it.kind === "triggered" && it.source_card_id === src.instance_id,
    ) ?? null
  );
}

function findCardAnywhere(v: SnapshotView, name: string): SnapshotCard | null {
  for (const seat of v.seats) {
    for (const z of [seat.hand, seat.graveyard, seat.library, seat.command]) {
      const c = findCardInZone(z, name);
      if (c) return c;
    }
  }
  return findCardInZone(v.exile, name);
}

// resolveStack passes priority around the table — through each
// player's own browser, by clicking the "next" button in whichever
// seat currently holds priority — until stack_items is empty. Going
// through the UI (rather than an admin pass_priority) keeps the
// server's holds-priority gate in play, so a browser that already
// auto-passed can't be double-passed into a step advance.
//
// Priority can move under us between reading the admin snapshot
// and clicking (a browser auto-pass, a no-priority step the engine
// walks through on its own), so every click is short-timeout and
// best-effort: a click that finds the button disabled just
// re-reads the snapshot and tries the new holder. Bounded at
// `maxAttempts` so a wedged stack fails loudly instead of spinning.
export async function resolveStack(setup: S19Setup, maxAttempts = 20): Promise<SnapshotView> {
  const { admin, caster, opponent } = setup;
  for (let i = 0; i < maxAttempts; i++) {
    const v = admin.snapshot();
    const items = v.stack_items ?? [];
    if (items.length === 0) return v;
    const holderSeat = v.turn?.priority_holder ?? -1;
    const holder = v.seats.find((s) => s.seat === holderSeat);
    const page =
      holder?.id === caster.playerID
        ? caster.page
        : holder?.id === opponent.playerID
          ? opponent.page
          : null;
    if (!page) {
      // No-priority step (untap / cleanup) — the engine advances
      // past it by itself; give it a beat and re-read.
      await new Promise((r) => setTimeout(r, 250));
      continue;
    }
    const before = items.length;
    try {
      await page.getByRole("button", { name: /^next$/ }).click({ timeout: 2000 });
    } catch {
      // Button went disabled under us — priority moved. Re-read.
      continue;
    }
    try {
      await admin.waitFor(
        (nv) =>
          (nv.stack_items ?? []).length < before ||
          (nv.turn?.priority_holder ?? -1) !== holderSeat,
        `priority rotates or stack shrinks (pass ${i + 1})`,
        5000,
      );
    } catch {
      // Fall through and re-evaluate from the latest snapshot.
    }
  }
  const last = admin.snapshot();
  throw new Error(
    `resolveStack: stack still non-empty after ${maxAttempts} attempts ` +
      `(step=${last.turn?.step} holder=${last.turn?.priority_holder} ` +
      `stack_items=${JSON.stringify((last.stack_items ?? []).map((it) => it.label ?? it.kind))})`,
  );
}

// keepAllHands has every seat keep its opening hand through the admin
// connection, in the order the server asks (CR 103.5, #2237): starting
// seat first, then round the table. A keep out of turn is refused, so
// the order cannot be hard-coded; it follows `mulligan_turn` instead.
export async function keepAllHands(admin: AdminClient): Promise<void> {
  for (let i = 0; i < 16; i++) {
    const v = await admin.waitFor(
      (s) => s.mulligans_open !== true || s.seats.some((p) => p.mulligan_turn === true),
      "a seat to decide on its opening hand",
      10_000,
    );
    if (v.mulligans_open !== true) return;
    const who = v.seats.find((p) => p.mulligan_turn === true)!;
    await admin.sendActionAsPlayer(who.id, "keep_hand", {});
    await admin.waitFor(
      (s) => s.mulligans_open !== true || s.seats.find((p) => p.mulligan_turn === true)?.id !== who.id,
      "the next seat to decide on its opening hand",
      10_000,
    );
  }
  throw new Error("the opening hands were not all kept after 16 decisions");
}

// setupS19Game spins up a 2-player game seeded with the S19 caster
// and opponent decks, walks both players through join, and uses the
// admin WS to fire keep_hand for both seats — the game is in
// StateActive with mulligans closed by the time this returns.
//
// No pre-setup throttle: playwright.config.ts starts the server with
// CMDCTRL_DEV_RELAX_RATE_LIMITS=1, which lifts the lobby's join/login
// rate limiter (5-burst, 1/second from one IP) that used to require a
// sleep between consecutive serial tests. If you attach the suite to
// an already-running dev server (reuseExistingServer), start it with
// that env var set, or tests 3+ may hit the limiter and time out on
// the join-page localStorage check.
//
// The admin layer drives keep_hand instead of the player UIs because
// the e2e suite focuses on trigger behaviour, not the mulligan modal
// (which has its own coverage in the lobby/full-game suite). It also
// sidesteps the post-Pixi-rewrite UI churn that broke the dialog-
// based flow.
export interface S19Options {
  // The gameplay settings each seat starts from, in place of
  // S19_GAMEPLAY. The stack-hold spec seats an opponent on the
  // defaults, which is the only way to watch smart autopass hold.
  casterGameplay?: Record<string, unknown>;
  opponentGameplay?: Record<string, unknown>;
  // A deck in place of the S19 caster's or opponent's (buildDeck in
  // s19-deck-fixture.ts). Still 100 cards.
  casterDeck?: string;
  opponentDeck?: string;
  // Both browsers' viewport, in place of Playwright's default.
  viewport?: { width: number; height: number };
}

export async function setupS19Game(
  browser: Browser,
  request: APIRequestContext,
  opts: S19Options = {},
): Promise<S19Setup> {
  const adminToken = await adminLogin(request);
  const game = await createGame(request, adminToken, `S19 e2e ${Date.now()}`);
  if (!game.invite_token) throw new Error("invite token missing on fresh game");

  const caster = await joinAsPlayer(
    browser,
    game.id,
    game.invite_token,
    "Caster",
    opts.casterGameplay,
    opts.viewport,
  );
  const opponent = await joinAsPlayer(
    browser,
    game.id,
    game.invite_token,
    "Opponent",
    opts.opponentGameplay,
    opts.viewport,
  );

  const up1 = await uploadDeckAs(
    request,
    adminToken,
    game.id,
    caster.playerID,
    opts.casterDeck ?? makeS19CasterDeck(),
  );
  const up2 = await uploadDeckAs(
    request,
    adminToken,
    game.id,
    opponent.playerID,
    opts.opponentDeck ?? makeS19OpponentDeck(),
  );
  if (up1.card_count !== 100) {
    throw new Error(`caster deck card_count=${up1.card_count}, want 100`);
  }
  if (up2.card_count !== 100) {
    throw new Error(`opponent deck card_count=${up2.card_count}, want 100`);
  }

  await startGameAs(request, adminToken, game.id);

  const admin = await openAdminClient(adminToken, game.id, caster.playerID, opponent.playerID);
  if (admin.snapshot().seats.length !== 2) {
    throw new Error("admin WS sees wrong seat count");
  }

  // Drive keep_hand for both seats from admin so the test doesn't
  // depend on the player-side mulligan dialog. Admin acts on behalf
  // of each player by populating `player` in the action payload.
  await keepAllHands(admin);

  // After both keep, the game state is active and mulligans_open
  // flips false. The hub may still be flushing snapshots — wait
  // until the admin sees the cleared state.
  await admin.waitFor((v) => v.state === "active", "game state active");

  // Both browsers auto-pass upkeep → draw → precombat main (the
  // active player's default stop). Wait for the cursor to settle
  // there before handing control to the test: staging a trigger
  // while a browser's pass_priority is still in flight lets that
  // pass rotate priority under the test's feet.
  await admin.waitFor(
    (v) => v.turn?.step === "precombat_main" && v.turn?.priority_holder === v.turn?.active_seat,
    "cursor settled on the active player's precombat main",
    15_000,
  );

  return {
    game,
    adminToken,
    admin,
    caster,
    opponent,
    async shutdown() {
      admin.close();
      await caster.context.close();
      await opponent.context.close();
    },
  };
}
