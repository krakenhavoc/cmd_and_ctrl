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
}

export interface PendingChoice {
  id: string;
  kind: string;
  chooser: string;
  source?: string;
  reason?: string;
}

export interface SnapshotView {
  state: string;
  seats: SnapshotPlayer[];
  battlefield: SnapshotZone;
  exile: SnapshotZone;
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
  const url = `ws://localhost:8080/ws?game=${gameID}&player=${encodeURIComponent(asSeatID)}&token=${encodeURIComponent(adminToken)}`;
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

// seedHandWithCard pumps `draw_card` for `ownerID` until `name`
// surfaces in their hand, then returns the matching SnapshotCard.
// Used by tests that need to know the hand size right BEFORE a
// trigger fires (the `adminMoveByName` shorthand bundles seeding
// and moving; this primitive splits them apart so tests can
// snapshot in between).
//
// The library / opponent-hand zones are wholesale-hidden by the
// per-viewer redactor (CR 400.2 private-zone rule), so admin can
// only address cards in a zone where they're identified by the
// per-viewer KnownBy machinery — the player's own hand qualifies
// because hand cards add their owner as a knower at deal time.
export async function seedHandWithCard(
  admin: AdminClient,
  ownerID: string,
  name: string,
): Promise<SnapshotCard> {
  const owner0 = playerByID(admin.snapshot(), ownerID);
  const already = findCardInZone(owner0.hand, name);
  if (already) return already;
  const startLibrary = owner0.library.count;
  const maxDraws = Math.min(startLibrary + 1, 105);
  let iters = 0;
  for (let i = 0; i < maxDraws; i++) {
    iters++;
    const beforeLib = playerByID(admin.snapshot(), ownerID).library.count;
    try {
      await admin.sendActionAsPlayer(ownerID, "draw_card", {});
    } catch (e) {
      throw new Error(
        `seedHandWithCard: draw_card rejected at iter ${iters}: ${(e as Error).message}`,
      );
    }
    // Wait for the library count to actually decrease — the merged
    // snapshot can lag a single broadcast frame when both admin
    // connections are racing to deliver the same broadcast on
    // separate sockets. waitFor pumps the merged-snapshot
    // subscribers until the library reflects the draw.
    try {
      await admin.waitFor(
        (v) => playerByID(v, ownerID).library.count < beforeLib,
        `caster library decreases past ${beforeLib}`,
        2000,
      );
    } catch {
      // No decrease within 2s — could be StepDraw early-return
      // or a no-op condition. Skip this iteration but don't fail
      // the whole loop.
      continue;
    }
    const owner = playerByID(admin.snapshot(), ownerID);
    const found = findCardInZone(owner.hand, name);
    if (found) return found;
    if (owner.library.count === 0) break;
  }
  const finalView = admin.snapshot();
  const ownerNow = playerByID(finalView, ownerID);
  const seatList = finalView.seats.map((s) => `${s.name}(${s.id.slice(0, 8)})`).join(",");
  const namesInHand = ownerNow.hand.cards.map((c) => c.name ?? "?");
  throw new Error(
    `seedHandWithCard: ${name} not surfaced for ${ownerID.slice(0, 8)} after ${iters} draws; ` +
      `seats=[${seatList}] ` +
      `library_count=${ownerNow.library.count} hand_count=${ownerNow.hand.count} ` +
      `graveyard_count=${ownerNow.graveyard.count} ` +
      `hand=${JSON.stringify(namesInHand)}`,
  );
}

// adminMoveByName routes a card identified by name to the target
// zone. The wire-side library / opponent-hand zones are wholesale-
// hidden by the per-viewer redactor (CR 400.2 private-zone rule),
// so admin can only "see" cards in the OWNING player's hand /
// graveyard / battlefield. To make a library card visible, this
// helper calls seedHandWithCard, then moves the resulting card
// from the hand to dstKind.
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
  let card: SnapshotCard | null = null;
  let srcKind: "hand" | "graveyard" | null = null;

  // Probe 1: already in hand?
  {
    const owner = playerByID(admin.snapshot(), ownerID);
    card = findCardInZone(owner.hand, name);
    if (card) srcKind = "hand";
    else {
      const grave = findCardInZone(owner.graveyard, name);
      if (grave) {
        card = grave;
        srcKind = "graveyard";
      }
    }
  }

  // Probe 2: seed it into the hand by drawing until found.
  if (!card) {
    card = await seedHandWithCard(admin, ownerID, name);
    srcKind = "hand";
  }

  const dst: { kind: string; owner?: string } = { kind: dstKind };
  if (dstKind === "graveyard" || dstKind === "hand") {
    dst.owner = ownerID;
  }
  const src: { kind: string; owner?: string } = { kind: srcKind, owner: ownerID };
  // Route through the owner's admin connection so action.Caller
  // matches the card's controller — the move_card gate
  // (`requireCardController`) only bypasses when Caller=uuid.Nil,
  // which a seat-bound admin WS isn't.
  return await admin.sendActionAsPlayer(ownerID, "move_card", {
    src,
    dst,
    instance_id: card.instance_id,
  });
}

// --- Browser orchestration -----------------------------------

export interface JoinedPlayer {
  context: BrowserContext;
  page: Page;
  name: string;
  token: string;
  playerID: string;
}

async function joinAsPlayer(
  browser: Browser,
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<JoinedPlayer> {
  const context = await browser.newContext();
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

// setupS19Game spins up a 2-player game seeded with the S19 caster
// and opponent decks, walks both players through join, and uses the
// admin WS to fire keep_hand for both seats — the game is in
// StateActive with mulligans closed by the time this returns.
//
// Throttled by a 2-second pre-setup sleep so the lobby's join
// rate limiter (5-burst, 1/second from one IP) refills between
// consecutive tests. Without this, tests 3+ in a serial run hit
// the limiter and the join-page localStorage check times out.
//
// The admin layer drives keep_hand instead of the player UIs because
// the e2e suite focuses on trigger behaviour, not the mulligan modal
// (which has its own coverage in the lobby/full-game suite). It also
// sidesteps the post-Pixi-rewrite UI churn that broke the dialog-
// based flow.
export async function setupS19Game(
  browser: Browser,
  request: APIRequestContext,
): Promise<S19Setup> {
  // Sleep long enough for the join rate limiter (1 token/s) to
  // refill the 4 tokens this test will consume (2 per browser
  // join + slop) so the suite stays green when running serially.
  await new Promise((r) => setTimeout(r, 4000));
  const adminToken = await adminLogin(request);
  const game = await createGame(request, adminToken, `S19 e2e ${Date.now()}`);
  if (!game.invite_token) throw new Error("invite token missing on fresh game");

  const caster = await joinAsPlayer(browser, game.id, game.invite_token, "Caster");
  const opponent = await joinAsPlayer(browser, game.id, game.invite_token, "Opponent");

  const up1 = await uploadDeckAs(request, adminToken, game.id, caster.playerID, makeS19CasterDeck());
  const up2 = await uploadDeckAs(request, adminToken, game.id, opponent.playerID, makeS19OpponentDeck());
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
  await admin.sendActionAsPlayer(caster.playerID, "keep_hand", {});
  await admin.sendActionAsPlayer(opponent.playerID, "keep_hand", {});

  // After both keep, the game state is active and mulligans_open
  // flips false. The hub may still be flushing snapshots — wait
  // until the admin sees the cleared state.
  await admin.waitFor((v) => v.state === "active", "game state active");

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
