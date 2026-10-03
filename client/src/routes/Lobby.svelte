<script lang="ts">
  import { onMount } from "svelte";
  import {
    addBotSeat,
    archiveGame,
    createGame,
    deleteGame,
    fetchBotOptions,
    fetchMySetup,
    fetchMyGames,
    getGame,
    joinByCode,
    joinGame,
    fetchTableSettings,
    listGames,
    mintSeatReclaim,
    rejoinMyGame,
    removeBotSeat,
    replayURL,
    patchTableSettings,
    rotateInvite,
    startGame,
    unarchiveGame,
    type BotOptions,
    type GameMeta,
    type ReclaimTicket,
    type SeatInfo,
  } from "../lib/api";
  import { inviteURL, reclaimURL, spectatorInviteURL, navigate } from "../lib/router";
  import { session, LobbyApiError } from "../lib/session";
  import { adminNotice, isAdmin as isAdminSession } from "../lib/admin";
  import { signedInUserID } from "../lib/myGames";
  import { GUEST_CODE_MESSAGE, inviteHash, joinBoxFor, myTables } from "../lib/signedInHome";
  import { seatColor } from "../lib/colors";
  import { avatarURL } from "../lib/api";
  import { canInviteTablemates } from "../lib/tablemates";
  import {
    canCreateTables,
    setupResultMessage,
    setupSummary,
    type TableSetup,
  } from "../lib/tableSetup";
  import type { TableSettingsView } from "../lib/protocol";
  import type { TableSettingsPatch } from "../lib/tableSettings";
  import DeckUploadForm from "../lib/components/DeckUploadForm.svelte";
  import TableSettingsPanel from "../lib/components/TableSettingsPanel.svelte";
  import TablematePicker from "../lib/components/TablematePicker.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";

  // Lobby is the signed-in home (ADR 0112 §1): every session lands
  // here, under the site header and its account menu. At the top, a
  // "Join a table" card (a code or a link for a signed-in person, a
  // link for a guest, none for the admin token) beside the create form
  // (any signed-in person, and the admin). Under them, the tables:
  // a signed-in person's own (myTables), every table for an admin.
  // Each table shows its seats, the deck-upload panel for your own
  // seat, and the way into the game once all seats are ready.
  let games = $state<GameMeta[]>([]);
  let error = $state("");
  let newName = $state("");
  let busy = $state(false);

  // --- creating a table, and the last setup (ADR 0110 §5) ---------
  //
  // Any signed-in person may create a table (owner answer 2). Their
  // last setup (the bots, settings and people of the last table they
  // created that started) is offered as "use my last setup", which the
  // server applies to the new table. Right after creating, the new
  // table opens its tablemate picker, people from the last setup
  // first.
  const canCreate = $derived(canCreateTables($session));

  // --- joining a table (ADR 0112 §1 item 3) ------------------------
  //
  // The box that used to be on the login page. A pasted link opens its
  // Join page, whoever is asking. A bare code is posted to POST /join
  // only by a signed-in person, whom the server seats as their Discord
  // identity; a guest is told before anything is sent, since the server
  // refuses a guest's code (409).
  const joinBox = $derived(joinBoxFor($session));
  let joinInput = $state("");
  let joining = $state(false);
  let joinError = $state("");

  async function onJoin(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    joinError = "";
    const raw = joinInput.trim();
    if (!raw) return;
    const hash = inviteHash(raw);
    if (hash) {
      navigate(hash);
      return;
    }
    if (joinBox !== "code") {
      joinError = GUEST_CODE_MESSAGE;
      return;
    }
    joining = true;
    try {
      // The new seat session replaces this page's session, so the list
      // below shows the new table with its deck panel, as the old
      // login-page flow did once it landed here.
      await joinByCode(raw);
      joinInput = "";
      await refresh();
    } catch (err) {
      joinError = err instanceof LobbyApiError ? err.message : "could not join with that code";
    } finally {
      joining = false;
    }
  }

  // --- your tables (ADR 0112 §1 item 6, owner answer 3) -------------
  //
  // The open tables where this person holds a seat through ANOTHER
  // session: GET /me/games entries with a `rejoin` path, by game id. A
  // table there that this session is not bound to gets an "Open" button,
  // which trades this session for that seat's (POST /me/games/{id}/
  // session) and then opens it.
  let rejoinPaths = $state<Map<string, string>>(new Map());
  const rejoinable = $derived(new Set(rejoinPaths.keys()));
  let opening = $state<string | null>(null);

  async function loadMyGames(): Promise<void> {
    if (!signedInUserID($session)) {
      rejoinPaths = new Map();
      return;
    }
    try {
      const mine = await fetchMyGames();
      const next = new Map<string, string>();
      for (const g of mine) if (g.rejoin) next.set(g.id, g.rejoin);
      rejoinPaths = next;
    } catch {
      // Keep what we had: a failed read must not hide a table, and the
      // list below still shows this session's own and created tables.
    }
  }

  // rejoinsHere: a table where this person holds a seat that this
  // session is not bound to.
  function rejoinsHere(g: GameMeta): boolean {
    return rejoinable.has(g.id) && $session?.gameID !== g.id;
  }

  async function openMine(g: GameMeta): Promise<void> {
    const path = rejoinPaths.get(g.id);
    if (!path || opening) return;
    opening = g.id;
    error = "";
    try {
      await rejoinMyGame(path);
      // A table still in the lobby opens here, where the deck is
      // imported; one underway opens on the board (as My games does).
      if (g.state === "lobby") await refresh();
      else navigate(`#/games/${g.id}`);
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "couldn't open that table";
    } finally {
      opening = null;
    }
  }
  let lastSetup = $state<TableSetup | null>(null);
  let useLastSetup = $state(true);
  // The table this page just created: its tablemate picker starts open.
  let justCreated = $state<string | null>(null);
  // What applying the setup did, shown on the new table's card.
  let createdNote = $state<{ gameID: string; message: string } | null>(null);
  let seatBusy = $state<string | null>(null);

  // Track the freshly-created game's invite tokens client-side — the
  // List endpoint strips both invites, so we remember them per
  // session so admins can copy each link without re-fetching
  // /games/{id}. The spectator invite (S11) is distinct from the
  // player invite — sharing the player one with a spectator would
  // let them claim a seat.
  const recentInvites = new Map<string, string>();
  const recentSpectatorInvites = new Map<string, string>();

  // --- bot seats (S31, ADR 0033 §9) -------------------------------
  //
  // Bots take REAL seats out of the same four, so a seated human can
  // add at most three. The control therefore lives on the open-seat
  // placeholders: when there is no open seat there is nothing to add
  // a bot to, and the affordance disappears on its own rather than
  // needing a count.
  //
  // The option catalog is game-independent — the same tiers and decks
  // for every table — so it is fetched once on mount. `enabled` false
  // means this server has no bot host at all; we hide the control
  // rather than offer a button that 503s.
  let botOptions = $state<BotOptions | null>(null);
  // The game whose Add-bot picker is open, or null.
  let botPickerFor = $state<string | null>(null);
  let botTier = $state("");
  let botDeck = $state("");
  let botBusy = $state(false);
  let botError = $state("");

  const botTiersAvailable = $derived((botOptions?.tiers ?? []).filter((t) => t.available));
  const botsOfferable = $derived(
    botOptions?.enabled === true &&
      botTiersAvailable.length > 0 &&
      (botOptions?.decks.length ?? 0) > 0,
  );

  // canManageBots: admin, or a player seated at this table. Mirrors
  // the server gate — the endpoint is authoritative, this only keeps
  // us from rendering a button that always 403s.
  function canManageBots(g: GameMeta): boolean {
    if (g.state !== "lobby" || !botsOfferable) return false;
    if (isAdmin) return true;
    return mySeat(g) !== null;
  }

  // --- table settings (ADR 0075 §2.5) -----------------------------
  //
  // The lobby half of the settings panel. Only the manager is offered
  // it here, and that is a LIMITATION rather than a decision: the
  // settings are public (every viewer gets them on the game view),
  // but the lobby's HTTP surface has only a write — PATCH — so the
  // read below is an empty patch and inherits the write's gate. At
  // the table, where the snapshot carries them, everyone sees the
  // same panel read-only. A `GET /games/{id}/settings` would close
  // the gap; it is server work and belongs in its own change.
  let settingsFor = $state<string | null>(null);
  let tableSettings = $state<TableSettingsView | null>(null);
  let settingsBusy = $state(false);
  let settingsError = $state<string | null>(null);

  function canManageTableFor(g: GameMeta): boolean {
    return isAdmin || mySeat(g)?.is_host === true;
  }

  async function openTableSettings(gameID: string): Promise<void> {
    settingsFor = gameID;
    tableSettings = null;
    settingsError = null;
    settingsBusy = true;
    try {
      tableSettings = await fetchTableSettings(gameID);
    } catch (e) {
      settingsError = e instanceof Error ? e.message : String(e);
    } finally {
      settingsBusy = false;
    }
  }

  async function applyTableSettings(gameID: string, patch: TableSettingsPatch): Promise<void> {
    settingsBusy = true;
    settingsError = null;
    try {
      // The response is the WHOLE settings object after the patch, so
      // the panel re-renders from the server's answer rather than
      // from what it hoped it had set.
      tableSettings = await patchTableSettings(gameID, patch);
    } catch (e) {
      settingsError = e instanceof Error ? e.message : String(e);
    } finally {
      settingsBusy = false;
    }
  }

  // canRotateInvites: admin, or the table's own creator (#1098).
  // g.is_creator is computed server-side per viewer — see
  // redactMetaFor in server/internal/lobby/http.go — so this mirrors
  // the server's CanRotateInvites without ever seeing a raw creator
  // id; the endpoint stays authoritative either way.
  function canRotateInvites(g: GameMeta): boolean {
    return isAdmin || g.is_creator === true;
  }

  function openBotPicker(gameID: string): void {
    botError = "";
    botPickerFor = gameID;
    botTier = botTiersAvailable[0]?.tier ?? "";
    botDeck = botOptions?.decks[0]?.id ?? "";
  }

  async function onAddBot(gameID: string): Promise<void> {
    if (!botTier || !botDeck) return;
    botBusy = true;
    botError = "";
    try {
      await addBotSeat(gameID, { tier: botTier, deck: botDeck });
      botPickerFor = null;
      await refresh();
    } catch (err) {
      botError = err instanceof LobbyApiError ? err.message : "could not add the bot";
    } finally {
      botBusy = false;
    }
  }

  async function onRemoveBot(gameID: string, playerID: string): Promise<void> {
    botBusy = true;
    botError = "";
    try {
      await removeBotSeat(gameID, playerID);
      await refresh();
    } catch (err) {
      botError = err instanceof LobbyApiError ? err.message : "could not remove the bot";
    } finally {
      botBusy = false;
    }
  }

  function botDeckName(id: string | undefined): string {
    if (!id) return "";
    return botOptions?.decks.find((d) => d.id === id)?.name ?? id;
  }

  // The curated deck names are flavour ("Raid and Ransack", "Body
  // Count"), so the <option> text alone does not say which one
  // attacks. The description leads with the archetype, and it belongs
  // on the page rather than in a title= tooltip — a tooltip is not an
  // answer on a touch device, and picking an archetype is the whole
  // decision this control exists for.
  const botDeckDescription = $derived(
    botOptions?.decks.find((d) => d.id === botDeck)?.description ?? "",
  );

  // --- admin: retiring old tables ---------------------------------
  //
  // Two actions, deliberately not the same button. ARCHIVE hides the
  // table and keeps everything (snapshot, replay, invite tokens), so
  // it is the one to reach for when clearing the lobby down. DELETE
  // reaps the replay JSONL and the restore point along with the
  // table and cannot be undone, so it lives behind the archived view
  // and its own confirmation.
  // The shared token, or a signed-in person on the server's admin
  // allowlist (ADR 0110 §3, lib/admin.ts). The server decides every
  // admin route; this only decides what the lobby offers.
  const isAdmin = $derived(isAdminSession($session));
  let showArchived = $state(false);
  let archived = $state<GameMeta[]>([]);
  let manageBusy = $state(false);
  // The pending confirmation, if any: which table and which action.
  let confirming = $state<{ id: string; kind: "archive" | "delete" } | null>(null);

  // Tables whose invites this page has already asked the server for,
  // so a creator who reloads gets their links back (GET /games/{id}
  // serves them to the creator, in the process that minted them) and a
  // table whose plaintext is gone is not asked about on every poll.
  const askedInvites = new Set<string>();

  async function loadCreatorInvites(list: GameMeta[]): Promise<void> {
    for (const g of list) {
      if (!g.is_creator || recentInvites.has(g.id) || askedInvites.has(g.id)) continue;
      askedInvites.add(g.id);
      try {
        const full = await getGame(g.id);
        if (full.invite_token) recentInvites.set(g.id, full.invite_token);
        if (full.spectator_invite) recentSpectatorInvites.set(g.id, full.spectator_invite);
      } catch {
        // No links on this card; "new invite link" still works.
      }
    }
  }

  async function refresh(): Promise<void> {
    try {
      const [list] = await Promise.all([listGames(), loadMyGames()]);
      await loadCreatorInvites(list);
      games = list;
      if (isAdmin && showArchived) archived = await listGames({ archived: true });
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "list failed";
    }
  }

  async function toggleArchived(): Promise<void> {
    showArchived = !showArchived;
    if (showArchived) {
      try {
        archived = await listGames({ archived: true });
      } catch (err) {
        error = err instanceof LobbyApiError ? err.message : "could not list archived tables";
      }
    }
  }

  // manage runs one admin table action with shared busy / error /
  // confirmation handling, so the three of them cannot drift.
  async function manage(fn: () => Promise<unknown>, failure: string): Promise<void> {
    manageBusy = true;
    error = "";
    try {
      await fn();
      confirming = null;
      await refresh();
      if (showArchived) archived = await listGames({ archived: true });
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : failure;
    } finally {
      manageBusy = false;
    }
  }

  // --- admin: seat reclaim links ----------------------------------
  //
  // A reclaim link is a bearer credential for ONE player's seat:
  // whoever opens it plays as that player, hand and all. So it is
  // minted on demand (never derivable from the game or seat), shown
  // once, single use, and short-lived — and the panel says all three
  // out loud, because the person clicking "copy" is about to paste it
  // into a chat window.
  let reclaim = $state<{ gameID: string; ticket: ReclaimTicket; url: string } | null>(null);
  let reclaimBusy = $state<string | null>(null);
  // Scoped to the table it happened on — the seat rows live inside
  // the per-game {#each}, so an unscoped message would print under
  // every table on the page.
  let reclaimError = $state<{ gameID: string; message: string } | null>(null);

  async function onSeatLink(gameID: string, p: SeatInfo): Promise<void> {
    reclaimBusy = p.player_id;
    reclaimError = null;
    reclaim = null;
    try {
      const ticket = await mintSeatReclaim(gameID, p.player_id);
      reclaim = { gameID, ticket, url: reclaimURL(gameID, ticket.ticket) };
    } catch (err) {
      reclaimError = {
        gameID,
        message: err instanceof LobbyApiError ? err.message : "could not generate a seat link",
      };
    } finally {
      reclaimBusy = null;
    }
  }

  // Minutes rather than seconds: the TTL is a promise to the host
  // ("this stops working soon"), not a countdown to watch.
  function ttlLabel(t: ReclaimTicket): string {
    const mins = Math.max(1, Math.round(t.ttl_seconds / 60));
    const at = new Date(t.expires_at);
    const clock = Number.isNaN(at.getTime())
      ? ""
      : ` (until ${at.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })})`;
    return `${mins} min${clock}`;
  }

  async function onCreate(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!newName.trim()) return;
    busy = true;
    error = "";
    try {
      const withSetup = useLastSetup && lastSetup !== null;
      const meta = await createGame(newName.trim(), withSetup ? { setup: "last" } : {});
      if (meta.invite_token) recentInvites.set(meta.id, meta.invite_token);
      if (meta.spectator_invite) recentSpectatorInvites.set(meta.id, meta.spectator_invite);
      justCreated = meta.id;
      const note = setupResultMessage(meta.setup);
      createdNote = note ? { gameID: meta.id, message: note } : null;
      newName = "";
      await refresh();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "create failed";
    } finally {
      busy = false;
    }
  }

  async function onStart(id: string): Promise<void> {
    error = "";
    try {
      await startGame(id);
      await refresh();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "start failed";
    }
  }

  // Transient "✓ copied" flash keyed by "<gameID>:<kind>" so the two
  // copy buttons (player / spectator invite) on multiple game rows
  // each flash independently. Mirrors the Settings.svelte export-copy
  // pattern: status flash on success, surfaced error on failure
  // (writeText rejects on insecure origins and unfocused tabs —
  // silently swallowing that left users pasting nothing).
  let copied = $state<string | null>(null);
  const COPY_FLASH_MS = 1500;
  async function copyToClipboard(url: string, key: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(url);
      copied = key;
      setTimeout(() => {
        // Only clear if no newer copy came in meanwhile.
        if (copied === key) copied = null;
      }, COPY_FLASH_MS);
    } catch {
      error = "clipboard write failed — copy the link from a focused, HTTPS tab";
    }
  }

  async function copyInvite(id: string): Promise<void> {
    const token = recentInvites.get(id);
    if (!token) {
      // Can't happen from the UI — the button that calls this only
      // renders once a token is cached — but a stale token dropping
      // out from under a click is not a wall we need to hit. "New
      // link" (onRotateInvite below) is the actual way back in.
      error = "no invite token cached for this game — use “new link” to mint one";
      return;
    }
    await copyToClipboard(inviteURL(id, token), `${id}:invite`);
  }

  async function copySpectatorInvite(id: string): Promise<void> {
    const token = recentSpectatorInvites.get(id);
    if (!token) {
      error = "no spectator invite cached for this game — use “new link” to mint one";
      return;
    }
    await copyToClipboard(spectatorInviteURL(id, token), `${id}:spectator`);
  }

  // --- admin: rotating a lost or compromised invite ----------------
  //
  // Only the process that minted a game can show its invite
  // plaintext (ADR 0051 decision 4) — after a restart there is no way
  // to recover a lost link for a table that's still running, and
  // "re-open as admin" (the old hint) never actually worked. Rotating
  // is the real fix: it revokes the current invite of one kind and
  // mints a replacement, which works whether or not this process ever
  // held the old plaintext. The trade the UI has to be honest about:
  // the OLD link of that kind stops working the instant this
  // succeeds, so it asks first.
  type InviteKind = "player" | "spectator";
  let confirmRotate = $state<{ gameID: string; kind: InviteKind } | null>(null);
  let rotateBusy = $state<string | null>(null); // `${gameID}:${kind}`
  let rotateError = $state<{ gameID: string; message: string } | null>(null);
  // The freshly-minted link, shown once so the admin can copy or send
  // it — mirrors the seat-reclaim panel below.
  let rotated = $state<{ gameID: string; kind: InviteKind; url: string } | null>(null);

  function inviteKindLabel(kind: InviteKind): string {
    return kind === "player" ? "player invite" : "spectator link";
  }

  async function onRotateInvite(gameID: string, kind: InviteKind): Promise<void> {
    const key = `${gameID}:${kind}`;
    rotateBusy = key;
    rotateError = null;
    try {
      const res = await rotateInvite(gameID, kind);
      const url =
        kind === "player" ? inviteURL(gameID, res.token) : spectatorInviteURL(gameID, res.token);
      if (kind === "player") recentInvites.set(gameID, res.token);
      else recentSpectatorInvites.set(gameID, res.token);
      rotated = { gameID, kind, url };
      confirmRotate = null;
      // Forces the "copy invite" / "spectator link" buttons (which
      // key off recentInvites/recentSpectatorInvites, not reactive
      // state) to notice the map changed — same trick onCreate uses.
      await refresh();
    } catch (err) {
      rotateError = {
        gameID,
        message:
          err instanceof LobbyApiError
            ? err.message
            : `could not mint a new ${inviteKindLabel(kind)}`,
      };
    } finally {
      rotateBusy = null;
    }
  }

  // takeSeat sits the creator down at their own table, as themselves:
  // the server takes the seat's name from their Discord sign-in. The
  // new seat session replaces this page's session, so the page then
  // shows the table as a seated player would.
  async function takeSeat(g: GameMeta): Promise<void> {
    const invite = recentInvites.get(g.id);
    if (!invite) return;
    seatBusy = g.id;
    error = "";
    try {
      await joinGame(g.id, invite, "");
      await refresh();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "could not take a seat";
    } finally {
      seatBusy = null;
    }
  }

  function openGame(id: string): void {
    navigate(`#/games/${id}`);
  }

  // mySeat returns the seat this user occupies in game g, or null if
  // the session isn't bound to a seat in g (admin viewing someone
  // else's game, or a RolePlayer viewing a different game entirely).
  function mySeat(g: GameMeta) {
    const s = $session;
    if (!s?.playerID || s.gameID !== g.id) return null;
    return g.players.find((p) => p.player_id === s.playerID) ?? null;
  }

  // startsFromSeat: a seated player who may start this table from the
  // lobby — its host, or its creator. Any seated player may (the server
  // allows it); the button is offered to the person running the table
  // so four seats do not race to press it.
  function startsFromSeat(g: GameMeta): boolean {
    const me = mySeat(g);
    return me !== null && (me.is_host === true || g.is_creator === true);
  }

  // canStart returns true when Start would succeed: at least 2 seats
  // and every seat has uploaded a real deck.
  function canStart(g: GameMeta): boolean {
    return g.state === "lobby" && g.players.length >= 2 && g.players.every((p) => p.deck_uploaded);
  }

  // startReason explains a disabled primary action in one line —
  // the admin's Start table, or a seated player's Enter table.
  function startReason(g: GameMeta): string {
    if (g.state !== "lobby") return "";
    const me = mySeat(g);
    if (me && !me.deck_uploaded) return "Upload your deck below";
    if (g.players.length < 2) return "Needs at least two seats";
    const pending = g.players.filter((p) => !p.deck_uploaded);
    if (pending.length === 1) return `Waiting on ${seatName(pending[0])}'s deck`;
    if (pending.length > 1) return `Waiting on ${pending.length} decks`;
    if (me && !startsFromSeat(g)) return "Waiting for the host to start";
    return "";
  }

  function seatName(p: SeatInfo): string {
    return p.display_name || p.name;
  }

  function initials(p: SeatInfo): string {
    return seatName(p).trim().slice(0, 1).toUpperCase() || "?";
  }

  // Commander tables seat four; pad the grid with open slots so the
  // card reads as a table rather than a list.
  const TABLE_SEATS = 4;
  function seatSlots(g: GameMeta): (SeatInfo | null)[] {
    const sorted = [...g.players].sort((a, b) => a.seat - b.seat);
    const slots: (SeatInfo | null)[] = [...sorted];
    while (slots.length < TABLE_SEATS) slots.push(null);
    return slots;
  }

  function timeAgo(iso: string): string {
    const ms = Date.now() - Date.parse(iso);
    if (!Number.isFinite(ms) || ms < 0) return "just now";
    const m = Math.floor(ms / 60_000);
    if (m < 1) return "just now";
    if (m < 60) return `${m}m ago`;
    const h = Math.floor(m / 60);
    if (h < 24) return `${h}h ago`;
    const d = Math.floor(h / 24);
    return `${d}d ago`;
  }

  // What the list shows (lib/signedInHome.ts's myTables): every table
  // for an admin, a signed-in person's own tables, and a guest's view
  // as it always was.
  const visibleGames = $derived(myTables(games, $session, { admin: isAdmin, rejoinable }));

  const summary = $derived.by(() => {
    const list = visibleGames;
    if (list.length === 0) return "No tables yet.";
    const live = list.filter((g) => g.state === "active").length;
    const waiting = list.filter((g) => g.state === "lobby").length;
    const parts = [`${list.length} table${list.length === 1 ? "" : "s"}`];
    if (live) parts.push(`${live} in progress`);
    if (waiting) parts.push(`${waiting} in the lobby`);
    return parts.join(" · ");
  });

  // Poll while any visible table is still in the lobby so seats,
  // decks and the admin's Start show up without a manual refresh
  // (the lobby has no WebSocket; the game route does).
  const POLL_MS = 5000;
  onMount(() => {
    void refresh();
    // The last setup, for "use my last setup". Signed-in only; a
    // failure just leaves the option off.
    if (signedInUserID($session)) {
      void fetchMySetup()
        .then((r) => (lastSetup = r.setup ?? null))
        .catch(() => undefined);
    }
    // Best-effort: a server without the bot routes leaves botOptions
    // null and the Add-bot control simply never appears.
    void fetchBotOptions()
      .then((o) => (botOptions = o))
      .catch(() => undefined);
    const t = setInterval(() => {
      if (visibleGames.some((g) => g.state === "lobby")) void refresh();
    }, POLL_MS);
    return () => clearInterval(t);
  });

  // A table sent us here (ADR 0112 §2 item 5): its connection needed
  // admin mode, which is off now. Shown once, then dismissed.
  let modeNotice = $state("");
  $effect(() => {
    const n = $adminNotice;
    if (!n) return;
    modeNotice = n;
    adminNotice.set("");
  });
</script>

<!-- The site header is the Lobby's only chrome (ADR 0112 §1 item 4):
     its account menu took over the command bar's session chip, settings
     and sign-out controls. -->
<SiteHeader />

<section class="lobby">
  <div class="head">
    <div>
      <h1 class="title" aria-label="cmd_and_ctrl · lobby">Tables</h1>
      <p class="sub">{summary}</p>
    </div>
  </div>

  {#if modeNotice}
    <div class="mode-notice" role="status">
      <span>{modeNotice}</span>
      <button type="button" class="ghost" onclick={() => (modeNotice = "")}>dismiss</button>
    </div>
  {/if}

  {#if joinBox !== "none" || canCreate}
    <!-- Join first: on a narrow screen the two cards stack, and joining
         is what most people arrive to do. -->
    <div class="starts">
      {#if joinBox !== "none"}
        <form class="start-card" onsubmit={onJoin}>
          <h2 class="panel-h">join a table</h2>
          <div class="create-row">
            <input
              class="mono"
              type="text"
              placeholder={joinBox === "code" ? "invite code or link" : "invite link"}
              aria-label={joinBox === "code" ? "invite code or link" : "invite link"}
              bind:value={joinInput}
            />
            <button type="submit" class="primary" disabled={joining || !joinInput.trim()}>
              {joining ? "…" : "join"}
            </button>
          </div>
          <p class="start-help">
            {#if joinBox === "code"}
              Paste the code from your pod's invite, or the whole link. You sit as your Discord
              name. A spectator link opens the table read-only.
            {:else}
              Paste the whole invite link. A spectator link opens the table read-only.
            {/if}
          </p>
          {#if joinError}
            <p class="error" role="alert">{joinError}</p>
          {/if}
        </form>
      {/if}
      {#if canCreate}
        <form class="start-card create" onsubmit={onCreate}>
          <h2 class="panel-h">create game</h2>
          <div class="create-row">
            <input
              type="text"
              placeholder="game name"
              aria-label="game name"
              bind:value={newName}
            />
            <button type="submit" class="primary" disabled={busy || !newName.trim()}>create</button>
          </div>
          {#if lastSetup}
            <label class="use-setup">
              <input type="checkbox" bind:checked={useLastSetup} />
              <span>
                Use my last setup
                <span class="setup-sum">{setupSummary(lastSetup, botDeckName)}</span>
              </span>
            </label>
          {/if}
        </form>
      {/if}
    </div>
  {/if}

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if visibleGames.length === 0}
    {#if joinBox === "code"}
      <!-- ADR 0112 §1 item 7: a signed-in person with no tables. -->
      <div class="empty-card">
        <p>
          You're not at a table yet. Paste an invite code above, or create a table and invite your
          pod.
        </p>
        <div class="empty-links">
          <a class="btn-link" href="#/practice"
            ><Icon name="robot" size={13} /> Practice against bots</a
          >
          <a class="btn-link" href="#/decks"><Icon name="library" size={13} /> Check a deck</a>
          {#if signedInUserID($session)}
            <a class="btn-link" href="#/my-games"><Icon name="scroll" size={13} /> My games</a>
          {/if}
        </div>
      </div>
    {:else}
      <p class="muted empty">
        {#if canCreate}
          Create a table, then send the invite link to your pod.
        {:else}
          You're not seated at a table yet — ask for an invite link.
        {/if}
      </p>
    {/if}
  {:else}
    <ul class="games">
      {#each visibleGames as g (g.id)}
        {@const seat = mySeat(g)}
        {@const reason = startReason(g)}
        <li class="tcard" class:mine={seat !== null}>
          <div class="trow">
            <div class="tid">
              <div class="tname">{g.name}</div>
              <div class="chips">
                {#if g.state === "active"}
                  <span class="chip live"><i class="dot"></i>In progress</span>
                {:else if g.state === "ended"}
                  <span class="chip ended">Ended</span>
                {:else}
                  <span class="chip">Lobby</span>
                {/if}
                <span class="meta">{g.players.length} of {TABLE_SEATS} seats</span>
                <span class="meta">·</span>
                <span class="meta">created {timeAgo(g.created_at)}</span>
              </div>
            </div>
            <div class="actions">
              <div class="arow">
                {#if recentInvites.has(g.id)}
                  <button class:on={copied === `${g.id}:invite`} onclick={() => copyInvite(g.id)}>
                    {#if copied === `${g.id}:invite`}
                      <Icon name="check" size={13} /> Invite copied
                    {:else}
                      <Icon name="link" size={13} /> copy invite
                    {/if}
                  </button>
                {/if}
                {#if recentSpectatorInvites.has(g.id)}
                  <button
                    class:on={copied === `${g.id}:spectator`}
                    onclick={() => copySpectatorInvite(g.id)}
                  >
                    {#if copied === `${g.id}:spectator`}
                      <Icon name="check" size={13} /> Link copied
                    {:else}
                      <Icon name="link" size={13} /> Spectator link
                    {/if}
                  </button>
                {/if}
                {#if canRotateInvites(g)}
                  <!-- Replaces the old "re-open as admin to recover"
                       hint, which never actually worked (the invite
                       plaintext only ever lived in the memory of the
                       process that minted it — see docs/lobby.md).
                       This mints a real replacement. Admin or the
                       table's own creator (#1098); see
                       canRotateInvites above. -->
                  <button
                    class="ghost"
                    title="mint a new player invite — the current one stops working immediately"
                    disabled={rotateBusy !== null}
                    onclick={() => (confirmRotate = { gameID: g.id, kind: "player" })}
                  >
                    <Icon name="undo" size={13} />
                    {rotateBusy === `${g.id}:player` ? "…" : "new invite link"}
                  </button>
                  <button
                    class="ghost"
                    title="mint a new spectator link — the current one stops working immediately"
                    disabled={rotateBusy !== null}
                    onclick={() => (confirmRotate = { gameID: g.id, kind: "spectator" })}
                  >
                    <Icon name="undo" size={13} />
                    {rotateBusy === `${g.id}:spectator` ? "…" : "new spectator link"}
                  </button>
                {/if}
                {#if canManageTableFor(g)}
                  <!-- ADR 0075 §2.5. Offered to the host and the admin
                       because this page can only READ the settings
                       through the write route (see openTableSettings).
                       -->
                  <button
                    class="ghost"
                    class:on={settingsFor === g.id}
                    aria-expanded={settingsFor === g.id}
                    title="undos, starting life, commander damage, bot speed, spawning"
                    onclick={() =>
                      settingsFor === g.id ? (settingsFor = null) : openTableSettings(g.id)}
                  >
                    <Icon name="gear" size={13} /> table settings
                  </button>
                {/if}
                <!-- Mirror the server's downloadReplay gate: admins may
                     pull the replay any time after the lobby phase, but
                     players get 403 until the game has ended (the JSONL
                     carries unfiltered hidden information mid-game). -->
                {#if g.state === "ended" || (isAdmin && g.state !== "lobby")}
                  {@const url = replayURL(g.id)}
                  {#if url}
                    <a class="btn-link" href={url} download={`${g.id}.jsonl`}>
                      <Icon name="draw" size={13} /> Replay
                    </a>
                  {/if}
                {/if}
                {#if isAdmin}
                  <button
                    class="ghost"
                    title="hide this table from the lobby — nothing is deleted"
                    disabled={manageBusy}
                    onclick={() => (confirming = { id: g.id, kind: "archive" })}
                  >
                    archive
                  </button>
                {/if}
                {#if rejoinsHere(g)}
                  <!-- ADR 0112 §1 item 6: a seat this person holds through
                       another session. Opening it trades this session for
                       that seat's. -->
                  <button
                    class="primary"
                    disabled={opening !== null}
                    title="open your seat at this table"
                    onclick={() => openMine(g)}
                  >
                    {opening === g.id ? "…" : "Open"}
                    <Icon name="chevronRight" size={13} />
                  </button>
                {:else if g.state === "lobby" && g.is_creator && !seat && recentInvites.has(g.id)}
                  <!-- ADR 0110 §5 item 4: the creator sits down at their
                       own table, as themselves. -->
                  <button disabled={seatBusy !== null} onclick={() => takeSeat(g)}>
                    {seatBusy === g.id ? "…" : "take a seat"}
                  </button>
                {/if}
                {#if rejoinsHere(g)}
                  <!-- Open (above) is the only way in. -->
                {:else if g.state === "lobby" && seat && !startsFromSeat(g)}
                  <!-- Seated players wait for the host; the button goes
                       live once the poll sees the table start. -->
                  <button class="primary" disabled title="waiting for the host to start">
                    enter table <Icon name="chevronRight" size={13} />
                  </button>
                {:else if g.state === "lobby"}
                  {#if !seat}
                    <button onclick={() => openGame(g.id)}>open table</button>
                  {/if}
                  {#if g.players.length >= 2}
                    <button
                      class="primary"
                      disabled={!canStart(g)}
                      title={reason || "start the game"}
                      onclick={() => onStart(g.id)}
                    >
                      start table <Icon name="chevronRight" size={13} />
                    </button>
                  {/if}
                {:else if g.state === "active"}
                  <button class="primary" onclick={() => openGame(g.id)}>
                    {seat ? "enter table" : "open table"}
                    <Icon name="chevronRight" size={13} />
                  </button>
                {:else}
                  <button class="ghost" onclick={() => openGame(g.id)}>open table</button>
                {/if}
              </div>
              {#if reason}
                <div class="reason">{reason}</div>
              {/if}
            </div>
          </div>

          <ul class="seats" aria-label="seats">
            {#each seatSlots(g) as p, i (p?.player_id ?? `open-${i}`)}
              {#if p}
                {@const avatar = avatarURL(p.discord_id, p.discord_avatar_hash)}
                <li
                  class="seat"
                  class:you={p.player_id === $session?.playerID}
                  class:bot={p.is_bot}
                >
                  <span class="sav" class:bot={p.is_bot} style:border-color={seatColor(p.seat)}>
                    {#if p.is_bot}
                      <i class="botmark" style:color={seatColor(p.seat)}>
                        <Icon name="robot" size={18} />
                      </i>
                    {:else if avatar}
                      <img src={avatar} alt="" />
                    {:else}
                      <i style:background={seatColor(p.seat)}>{initials(p)}</i>
                    {/if}
                  </span>
                  <div class="sinfo">
                    <div class="sname">
                      seat {p.seat + 1}: {seatName(p)}
                      {#if p.is_bot}<b class="botchip">bot</b>{/if}
                      {#if p.is_host}<b title="Table host: may manage the table alongside the admin"
                          >host</b
                        >{/if}
                      {#if p.player_id === $session?.playerID}<b>you</b>{/if}
                    </div>
                    {#if p.is_bot}
                      <div class="sstat">
                        <i class="dot ok"></i>{p.bot_tier || "bot"} · {botDeckName(p.bot_deck) ||
                          p.deck_name ||
                          "deck ready"}
                      </div>
                    {:else if p.deck_uploaded}
                      <div class="sstat"><i class="dot ok"></i>{p.deck_name || "deck ready"}</div>
                    {:else}
                      <div class="sstat pend">deck pending</div>
                    {/if}
                  </div>
                  {#if !p.is_bot && isAdmin}
                    <!-- Admin-only, per seat: the disconnected player's
                         way back in. Sits where the bot's remove
                         button sits, because it is the same kind of
                         thing — a per-seat operator action. -->
                    <button
                      class="seat-link"
                      title={`generate a one-time link that returns ${seatName(p)} to this seat`}
                      aria-label={`seat link for ${seatName(p)}`}
                      disabled={reclaimBusy !== null}
                      onclick={() => onSeatLink(g.id, p)}
                    >
                      {reclaimBusy === p.player_id ? "…" : "seat link"}
                    </button>
                  {/if}
                  {#if p.is_bot && canManageBots(g)}
                    <button
                      class="seat-x"
                      title="remove this bot"
                      aria-label={`remove ${seatName(p)}`}
                      disabled={botBusy}
                      onclick={() => onRemoveBot(g.id, p.player_id)}
                    >
                      <Icon name="x" size={12} />
                    </button>
                  {/if}
                </li>
              {:else}
                <li class="seat open">
                  <span class="sav open"></span>
                  <div class="sinfo">
                    <div class="sname dim">Open seat</div>
                    <div class="sstat dim">
                      {g.state === "lobby" ? "send the invite link" : "—"}
                    </div>
                  </div>
                  {#if canManageBots(g) && botPickerFor !== g.id}
                    <button class="seat-add" disabled={botBusy} onclick={() => openBotPicker(g.id)}>
                      add bot
                    </button>
                  {/if}
                </li>
              {/if}
            {/each}
          </ul>

          {#if settingsFor === g.id}
            <div class="tsettings">
              {#if tableSettings}
                <TableSettingsPanel
                  settings={tableSettings}
                  canManage={canManageTableFor(g)}
                  gameState={g.state}
                  busy={settingsBusy}
                  error={settingsError}
                  onpatch={(patch) => applyTableSettings(g.id, patch)}
                />
              {:else if settingsBusy}
                <p class="muted">loading settings…</p>
              {:else}
                <p class="error">{settingsError ?? "couldn't load this table's settings"}</p>
              {/if}
            </div>
          {/if}

          {#if confirming?.id === g.id}
            <div class="confirm" class:danger={confirming.kind === "delete"} role="alert">
              {#if confirming.kind === "archive"}
                <p>
                  <strong>Archive “{g.name}”?</strong>
                  It leaves this list and stops taking part — bots stop, and anyone still connected is
                  dropped. Nothing is deleted: the board, the replay and the invite links are all kept,
                  and you can restore it from <em>archived tables</em> below.
                </p>
              {:else}
                <p>
                  <strong>Delete “{g.name}” permanently?</strong>
                  This destroys the board, the restore point and the replay file. It cannot be undone.
                  Archive instead if you only want it off the list.
                </p>
              {/if}
              <div class="confirm-actions">
                <button
                  class="primary"
                  disabled={manageBusy}
                  onclick={() =>
                    manage(
                      () => (confirming?.kind === "delete" ? deleteGame(g.id) : archiveGame(g.id)),
                      confirming?.kind === "delete" ? "delete failed" : "archive failed",
                    )}
                >
                  {confirming.kind === "delete" ? "delete permanently" : "archive table"}
                </button>
                <button class="ghost" disabled={manageBusy} onclick={() => (confirming = null)}>
                  cancel
                </button>
              </div>
            </div>
          {/if}

          {#if confirmRotate?.gameID === g.id}
            <div class="confirm" role="alert">
              <p>
                <strong>Mint a new {inviteKindLabel(confirmRotate.kind)}?</strong>
                The current {inviteKindLabel(confirmRotate.kind)} link stops working the moment this happens
                — anyone still holding it will need the new one.
              </p>
              <div class="confirm-actions">
                <button
                  class="primary"
                  disabled={rotateBusy !== null}
                  onclick={() => onRotateInvite(g.id, confirmRotate!.kind)}
                >
                  {rotateBusy === `${g.id}:${confirmRotate.kind}` ? "minting…" : "mint new link"}
                </button>
                <button
                  class="ghost"
                  disabled={rotateBusy !== null}
                  onclick={() => (confirmRotate = null)}
                >
                  cancel
                </button>
              </div>
            </div>
          {/if}

          {#if rotated?.gameID === g.id}
            <div class="reclaim">
              <div class="rhead">
                <span class="panel-h">new {inviteKindLabel(rotated.kind)}</span>
                <button
                  class="ghost"
                  aria-label="dismiss the new link"
                  onclick={() => (rotated = null)}><Icon name="x" size={12} /></button
                >
              </div>
              <input
                class="rurl"
                type="text"
                readonly
                value={rotated.url}
                aria-label={`new ${inviteKindLabel(rotated.kind)}`}
                onfocus={(e) => e.currentTarget.select()}
              />
              <div class="ractions">
                <button
                  class:on={copied === `${g.id}:rotated`}
                  onclick={() => copyToClipboard(rotated!.url, `${g.id}:rotated`)}
                >
                  {#if copied === `${g.id}:rotated`}
                    <Icon name="check" size={13} /> Link copied
                  {:else}
                    <Icon name="link" size={13} /> copy new link
                  {/if}
                </button>
              </div>
              <p class="hint warn">
                The old {inviteKindLabel(rotated.kind)} link no longer works. Send this one out instead.
              </p>
            </div>
          {/if}
          {#if rotateError?.gameID === g.id}
            <div class="bot-error">{rotateError.message}</div>
          {/if}

          {#if reclaim?.gameID === g.id}
            <div class="reclaim">
              <div class="rhead">
                <span class="panel-h">seat link — {reclaim.ticket.player_name}</span>
                <button
                  class="ghost"
                  aria-label="dismiss the seat link"
                  onclick={() => (reclaim = null)}><Icon name="x" size={12} /></button
                >
              </div>
              <input
                class="rurl"
                type="text"
                readonly
                value={reclaim.url}
                aria-label="seat reclaim link"
                onfocus={(e) => e.currentTarget.select()}
              />
              <div class="ractions">
                <button
                  class:on={copied === `${g.id}:reclaim`}
                  onclick={() => copyToClipboard(reclaim!.url, `${g.id}:reclaim`)}
                >
                  {#if copied === `${g.id}:reclaim`}
                    <Icon name="check" size={13} /> Link copied
                  {:else}
                    <Icon name="link" size={13} /> copy seat link
                  {/if}
                </button>
                <span class="rttl">
                  works once · expires in {ttlLabel(reclaim.ticket)}
                </span>
              </div>
              <p class="hint warn">
                Send this to {reclaim.ticket.player_name} and nobody else. Anyone who opens it plays as
                them — their hand included. It stops working the moment it is used.
              </p>
            </div>
          {/if}
          {#if reclaimError?.gameID === g.id}
            <div class="bot-error">{reclaimError.message}</div>
          {/if}

          {#if botPickerFor === g.id}
            <div class="bot-picker">
              <label>
                <span>tier</span>
                <select bind:value={botTier} disabled={botBusy}>
                  {#each botOptions?.tiers ?? [] as t (t.tier)}
                    <option value={t.tier} disabled={!t.available} title={t.description}>
                      {t.label}{t.available ? "" : " — not built yet"}
                    </option>
                  {/each}
                </select>
              </label>
              <label>
                <span>deck</span>
                <select bind:value={botDeck} disabled={botBusy}>
                  {#each botOptions?.decks ?? [] as d (d.id)}
                    <option value={d.id} title={d.description}>{d.name}</option>
                  {/each}
                </select>
              </label>
              {#if botDeckDescription}
                <p class="deck-desc">{botDeckDescription}</p>
              {/if}
              <div class="bot-actions">
                <button
                  class="primary"
                  disabled={botBusy || !botTier || !botDeck}
                  onclick={() => onAddBot(g.id)}
                >
                  {botBusy ? "adding…" : "add bot"}
                </button>
                <button class="ghost" disabled={botBusy} onclick={() => (botPickerFor = null)}>
                  cancel
                </button>
              </div>
              {#if botError}<div class="bot-error">{botError}</div>{/if}
              <p class="hint">
                A bot takes a real seat, arrives with its deck already loaded, and starts playing
                when you press start.
              </p>
            </div>
          {:else if botError}
            <div class="bot-error">{botError}</div>
          {/if}

          {#if createdNote?.gameID === g.id}
            <p class="hint setup-note" role="status">{createdNote.message}</p>
          {/if}

          {#if g.state === "lobby" && canInviteTablemates($session, $session?.gameID, g.id, g.is_creator === true)}
            <!-- Open on the table this page just created (ADR 0110 §5
                 item 3): inviting people is the next thing to do. -->
            <details class="invite-picker" open={justCreated === g.id}>
              <summary>
                <span class="panel-h">invite a tablemate</span>
              </summary>
              <TablematePicker gameID={g.id} lastTable={lastSetup?.tablemates ?? []} />
            </details>
          {/if}

          {#if seat && g.state === "lobby" && $session?.playerID}
            <details class="deck-upload" open={!seat.deck_uploaded}>
              <summary>
                <span class="panel-h">
                  {seat.deck_uploaded ? "replace your deck" : "choose your deck"}
                </span>
                {#if seat.deck_uploaded}
                  <span class="deck-ok"><i class="dot ok"></i>{seat.deck_name || "deck ready"}</span
                  >
                {/if}
              </summary>
              <p class="hint">
                Every deck is validated against Commander rules — 100-card singleton, color
                identity, format legality — whichever way it arrives.
              </p>
              <DeckUploadForm
                gameID={g.id}
                playerID={$session.playerID}
                onSuccess={() => void refresh()}
              />
            </details>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  {#if isAdmin}
    <section class="archive">
      <button class="ghost arch-toggle" onclick={toggleArchived}>
        {showArchived ? "hide" : "show"} archived tables
        {#if showArchived && archived.length > 0}({archived.length}){/if}
      </button>
      {#if showArchived}
        {#if archived.length === 0}
          <p class="muted empty">Nothing archived.</p>
        {:else}
          <ul class="arch-list">
            {#each archived as g (g.id)}
              <li class="arch-row">
                <div class="arch-id">
                  <span class="arch-name">{g.name}</span>
                  <span class="meta">
                    {g.players.length} seats · {g.state} · archived {g.archived_at
                      ? timeAgo(g.archived_at)
                      : ""}
                  </span>
                </div>
                <div class="arow">
                  <button
                    disabled={manageBusy}
                    onclick={() => manage(() => unarchiveGame(g.id), "restore failed")}
                  >
                    restore
                  </button>
                  <button
                    class="ghost danger"
                    disabled={manageBusy}
                    onclick={() => (confirming = { id: g.id, kind: "delete" })}
                  >
                    delete…
                  </button>
                </div>
                {#if confirming?.id === g.id && confirming.kind === "delete"}
                  <div class="confirm danger" role="alert">
                    <p>
                      <strong>Delete “{g.name}” permanently?</strong>
                      This destroys the board, the restore point and the replay file. It cannot be undone.
                    </p>
                    <div class="confirm-actions">
                      <button
                        class="primary"
                        disabled={manageBusy}
                        onclick={() => manage(() => deleteGame(g.id), "delete failed")}
                      >
                        delete permanently
                      </button>
                      <button
                        class="ghost"
                        disabled={manageBusy}
                        onclick={() => (confirming = null)}
                      >
                        cancel
                      </button>
                    </div>
                  </div>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      {/if}
    </section>
  {/if}

  <p class="foot">
    <button class="ghost" onclick={refresh}><Icon name="undo" size={13} /> refresh</button>
  </p>
</section>

<style>
  .lobby {
    max-width: 1080px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }
  .head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 20px;
    flex-wrap: wrap;
  }
  h1.title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 26px;
    font-weight: 800;
    letter-spacing: -0.02em;
    color: var(--fg);
    text-transform: none;
    line-height: 1.1;
  }
  .sub {
    margin: 4px 0 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .panel-h {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    font-weight: 600;
  }
  /* The two start cards (ADR 0112 §1 item 3): join and create, side by
     side, stacking join-first on a narrow screen. */
  .starts {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(320px, 100%), 1fr));
    gap: 12px;
  }
  .start-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
    padding: 14px 16px;
    border-radius: 12px;
    border: 1px solid var(--border);
    background: var(--surface);
  }
  .start-help {
    margin: 0;
    font-size: 12px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .create-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .create-row input[type="text"] {
    flex: 1;
    min-width: 0;
    margin: 0;
    height: 36px;
    padding: 0 12px;
    box-sizing: border-box;
    font-size: 13px;
  }
  .create-row input.mono {
    font-family: var(--font-mono);
    font-size: 12.5px;
  }
  .create-row button {
    height: 36px;
    flex: 0 0 auto;
  }
  .use-setup {
    display: flex;
    align-items: flex-start;
    gap: 6px;
    font-size: 12.5px;
    color: var(--fg);
    max-width: 420px;
  }
  .use-setup input {
    margin: 2px 0 0;
  }
  .setup-sum {
    display: block;
    font-size: 11.5px;
    color: var(--fg-muted);
  }
  .setup-note {
    margin: 8px 0 0;
  }
  .mode-notice {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin: 0 0 16px;
    padding: 10px 14px;
    border: 1px solid var(--border-strong);
    border-left: 3px solid var(--accent);
    border-radius: var(--radius);
    background: var(--surface);
    font-size: 13.5px;
  }
  .error {
    color: var(--danger);
    margin: 0;
    font-size: 13px;
  }
  .muted {
    color: var(--fg-muted);
  }
  .empty {
    margin: 12px 0;
    font-size: 13.5px;
  }
  /* A signed-in person's empty Lobby (ADR 0112 §1 item 7). */
  .empty-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 18px 20px;
    border-radius: 12px;
    border: 1px dashed var(--border-strong);
  }
  .empty-card p {
    margin: 0;
    font-size: 13.5px;
    color: var(--fg);
  }
  .empty-links {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .games {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .tcard {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 18px 22px 20px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .tcard.mine {
    border-color: rgba(217, 180, 92, 0.3);
    box-shadow: inset 0 0 0 1px rgba(217, 180, 92, 0.08);
  }
  .trow {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 20px;
    flex-wrap: wrap;
  }
  .tid {
    min-width: 0;
  }
  .tname {
    font-family: var(--font-display);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--fg);
  }
  .chips {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 8px;
    flex-wrap: wrap;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    height: 20px;
    padding: 0 8px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
    font-weight: 600;
  }
  .chip.live {
    color: var(--gold-strong);
    border-color: rgba(217, 180, 92, 0.5);
    background: var(--gold-soft);
  }
  .chip.ended {
    color: var(--fg-dim);
    border-color: var(--border);
  }
  .dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--gold);
    flex: 0 0 auto;
  }
  .dot.ok {
    background: var(--mint);
  }
  .meta {
    font-family: var(--font-mono);
    font-size: 10.5px;
    color: var(--fg-dim);
    letter-spacing: 0.04em;
  }
  .actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 8px;
    flex: 0 0 auto;
  }
  .arow {
    display: flex;
    gap: 6px;
    align-items: center;
    flex-wrap: wrap;
    justify-content: flex-end;
  }
  .arow button,
  .btn-link {
    height: 32px;
    padding: 0 12px;
    font-size: 12.5px;
  }
  .arow button.on {
    color: var(--mint);
    border-color: rgba(95, 212, 164, 0.4);
  }
  .btn-link {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    color: var(--fg);
    font-weight: 600;
    text-decoration: none;
    box-sizing: border-box;
  }
  .btn-link:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .reason {
    font-size: 11px;
    color: var(--fg-dim);
    text-align: right;
  }

  .seats {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px;
  }
  @media (max-width: 760px) {
    .seats {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
  .seat {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 12px;
    border-radius: 10px;
    background: var(--surface-sunken);
    border: 1px solid transparent;
    min-width: 0;
  }
  .seat.you {
    border-color: rgba(217, 180, 92, 0.4);
  }
  .seat.open {
    border: 1px dashed var(--border-strong);
    background: transparent;
  }
  .sav {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 2px solid var(--border-strong);
    padding: 2px;
    box-sizing: border-box;
    background: var(--surface-raised);
    flex: 0 0 auto;
    overflow: hidden;
  }
  .sav img,
  .sav i {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    border-radius: 50%;
    object-fit: cover;
    font-style: normal;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 14px;
    color: #1c1503;
  }
  .sav.open {
    border: 2px dashed var(--border-strong);
    background: transparent;
  }
  .sinfo {
    min-width: 0;
  }
  .sname {
    font-size: 12.5px;
    font-weight: 700;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .sname b {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--gold-strong);
    font-weight: 700;
    margin-left: 5px;
  }
  .sname.dim,
  .sstat.dim {
    color: var(--fg-dim);
    font-weight: 500;
  }
  .sstat {
    font-size: 11px;
    color: var(--fg-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-top: 1px;
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .sstat.pend {
    color: var(--gold-strong);
  }

  /* --- bot seats (S31) ------------------------------------------- */
  .seat.bot {
    border-style: dashed;
    border-color: var(--border-strong);
  }
  .sav.bot {
    border-style: dashed;
  }
  .sav .botmark {
    background: transparent;
    color: inherit;
  }
  .sname b.botchip {
    color: var(--fg-muted);
  }
  /* The seat row is a flex row; these push to its trailing edge. */
  .seat-x,
  .seat-add,
  .seat-link {
    margin-left: auto;
    flex: 0 0 auto;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: transparent;
    color: var(--fg-muted);
    cursor: pointer;
    font: inherit;
  }
  .seat-x {
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    padding: 0;
  }
  .seat-add,
  .seat-link {
    font-size: 11px;
    padding: 4px 9px;
    white-space: nowrap;
  }
  .seat-x:hover:not(:disabled),
  .seat-add:hover:not(:disabled),
  .seat-link:hover:not(:disabled) {
    color: var(--fg);
    border-color: var(--gold);
  }
  .seat-x:disabled,
  .seat-add:disabled,
  .seat-link:disabled {
    opacity: 0.5;
    cursor: default;
  }

  /* --- admin table management ------------------------------------ */
  /* Same card-inset shell as .confirm, so a disclosure inside a table
     card reads the same whether it is a settings panel or a
     confirmation. */
  .tsettings {
    border: 1px solid var(--border-strong);
    border-radius: 10px;
    padding: 12px 14px;
    background: var(--surface-sunken);
  }
  .confirm {
    border: 1px solid var(--border-strong);
    border-radius: 10px;
    padding: 12px 14px;
    background: var(--surface-sunken);
  }
  .confirm.danger {
    border-color: rgba(226, 96, 96, 0.45);
  }
  .confirm p {
    margin: 0;
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .confirm strong {
    color: var(--fg);
  }
  .confirm-actions {
    display: flex;
    gap: 8px;
    margin-top: 10px;
  }
  .arch-toggle {
    font-size: 12px;
  }
  .archive {
    border-top: 1px solid var(--border);
    padding-top: 14px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    align-items: flex-start;
  }
  .arch-list {
    list-style: none;
    margin: 0;
    padding: 0;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .arch-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 10px 14px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface-sunken);
  }
  .arch-id {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  .arch-name {
    font-family: var(--font-display);
    font-size: 14px;
    font-weight: 700;
    color: var(--fg-muted);
  }
  .arch-row .confirm {
    flex: 1 0 100%;
  }
  button.danger {
    color: var(--danger);
  }

  /* --- seat reclaim link ----------------------------------------- */
  .reclaim {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 14px;
    border: 1px solid rgba(217, 180, 92, 0.35);
    border-radius: 10px;
    background: var(--gold-soft);
  }
  .rhead {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .rhead button {
    width: 24px;
    height: 24px;
    padding: 0;
    display: grid;
    place-items: center;
  }
  .rurl {
    width: 100%;
    box-sizing: border-box;
    margin: 0;
    height: 32px;
    padding: 0 10px;
    font-family: var(--font-mono);
    font-size: 11.5px;
  }
  .ractions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }
  .ractions button {
    height: 30px;
    padding: 0 12px;
    font-size: 12px;
  }
  .ractions button.on {
    color: var(--mint);
    border-color: rgba(95, 212, 164, 0.4);
  }
  .rttl {
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
  }
  .hint.warn {
    margin: 0;
    color: var(--fg);
  }
  .bot-picker {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 10px;
    padding: 12px;
    border: 1px dashed var(--border-strong);
    border-radius: 10px;
    background: var(--surface-sunken);
  }
  .bot-picker label {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 11px;
    color: var(--fg-muted);
    min-width: 150px;
  }
  .bot-picker select {
    font: inherit;
    font-size: 12.5px;
    padding: 6px 8px;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    color: var(--fg);
  }
  .bot-actions {
    display: flex;
    gap: 8px;
  }
  .bot-picker .hint {
    flex: 1 0 100%;
    margin: 0;
  }
  /* Breaks the picker row so the description reads as a caption under
     the two selects rather than a third column squeezed beside them. */
  .deck-desc {
    flex: 1 0 100%;
    margin: 0;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .bot-error {
    color: var(--danger);
    font-size: 12px;
    flex: 1 0 100%;
  }

  /* The invite picker wears the deck panel's disclosure exactly, so
     a game card reads as one stack of panels rather than two
     unrelated controls (ADR 0051 decision 8, S34 sub-PR 6). */
  .deck-upload,
  .invite-picker {
    border-top: 1px solid var(--border);
    padding-top: 12px;
  }
  .invite-picker summary,
  .deck-upload summary {
    cursor: pointer;
    list-style: none;
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .invite-picker summary::-webkit-details-marker,
  .deck-upload summary::-webkit-details-marker {
    display: none;
  }
  .invite-picker summary::before,
  .deck-upload summary::before {
    content: "";
    width: 6px;
    height: 6px;
    border-right: 1.5px solid var(--fg-dim);
    border-bottom: 1.5px solid var(--fg-dim);
    transform: rotate(-45deg);
    transition: transform 120ms var(--ease);
  }
  .invite-picker[open] summary::before,
  .deck-upload[open] summary::before {
    transform: rotate(45deg);
  }
  .invite-picker summary:hover .panel-h,
  .deck-upload summary:hover .panel-h {
    color: var(--fg);
  }
  .deck-ok {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .hint {
    font-size: 12px;
    color: var(--fg-muted);
    line-height: 1.5;
    margin: 10px 0 0;
  }
  .foot {
    margin: 0;
  }
</style>
