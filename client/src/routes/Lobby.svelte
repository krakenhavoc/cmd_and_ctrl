<script lang="ts">
  import { onMount, tick } from "svelte";
  import { SvelteMap } from "svelte/reactivity";
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
    removeSeat,
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
  import { inviteURL, reclaimURL, spectatorInviteURL, navigate, route } from "../lib/router";
  import { L } from "../lib/labels";
  import NewTableRow from "../lib/components/NewTableRow.svelte";
  import { session, LobbyApiError } from "../lib/session";
  import { lobbyErrorMessage } from "../lib/serverFull";
  import { adminNotice, isAdmin as isAdminSession } from "../lib/admin";
  import { signedInUserID, type MyGame } from "../lib/myGames";
  import { community } from "../lib/community";
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
  import {
    contextLine,
    filterHistory,
    HISTORY_FILTERS,
    HISTORY_FIRST,
    HISTORY_ID,
    HISTORY_STEP,
    historyRows,
    moreCount,
    nowTables,
    seatLine,
    seatName,
    statusChip,
    whenLabel,
    type HistoryFilter,
    type HistoryRow,
  } from "../lib/tablesPage";
  import type { TableSettingsView } from "../lib/protocol";
  import type { TableSettingsPatch } from "../lib/tableSettings";
  import DeckUploadForm from "../lib/components/DeckUploadForm.svelte";
  import TableSettingsPanel from "../lib/components/TableSettingsPanel.svelte";
  import TablematePicker from "../lib/components/TablematePicker.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import MoreMenu, { type MoreItem } from "../lib/components/MoreMenu.svelte";
  import RemovedNotice from "../lib/components/RemovedNotice.svelte";
  import { canRemoveSeat, removeSeatConsequence, removeSeatQuestion } from "../lib/seatRemoval";
  import { BLOCK_HINT, BLOCK_LABEL, canManageBlocks } from "../lib/moderation";
  import BlockList from "../lib/components/BlockList.svelte";
  import AgentChip from "../lib/components/AgentChip.svelte";
  import Page from "../lib/components/Page.svelte";
  import { noteMyGames } from "../lib/hints/endedGame";
  import { defaultBotTier, tierNote, tierOptionLabel } from "../lib/botTier";

  // Tables, the signed-in home (ADR 0112 §1, ADR 0158 §1). The page
  // head holds the join box and + New table. Under it, two sections:
  //
  //   - Now: the tables in the lobby or in play that this person sits
  //     at or created (every table for an admin), one card each with a
  //     primary button, Copy invite, a ⋯ menu, the four seats and the
  //     lobby's own controls in the footer.
  //   - History: every table they sat at that has ended or been
  //     archived (GET /me/games minus Now), one compact row each, with
  //     client-side filters. My games used to be its own page.
  let games = $state<GameMeta[]>([]);
  let error = $state("");
  let newName = $state("");
  let busy = $state(false);
  let loaded = $state(false);

  // --- creating a table, and the last setup (ADR 0110 §5) ---------
  //
  // Any signed-in person may create a table (owner answer 2). Their
  // last setup is offered as "Use my last setup", which the server
  // applies to the new table. The form opens from + New table and stays
  // open after a failed create, so the error can be read.
  const canCreate = $derived(canCreateTables($session));
  let newOpen = $state(false);
  let createError = $state("");

  async function openNewTable(): Promise<void> {
    newOpen = !newOpen;
    createError = "";
    if (newOpen) {
      await tick();
      document.querySelector<HTMLInputElement>('input[aria-label="game name"]')?.focus();
    }
  }

  // --- joining a table (ADR 0112 §1 item 3) ------------------------
  //
  // A pasted link opens its Join page, whoever is asking. A bare code
  // is posted to POST /join only by a signed-in person, whom the server
  // seats as their Discord identity; a guest is told before anything is
  // sent, since the server refuses a guest's code (409).
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
      // The new seat session replaces this page's session, so Now
      // shows the new table with its deck panel.
      await joinByCode(raw);
      joinInput = "";
      await refresh();
    } catch (err) {
      joinError = lobbyErrorMessage(err, "could not join with that code");
    } finally {
      joining = false;
    }
  }

  // --- your games (GET /me/games) ----------------------------------
  //
  // Every seat this person holds, from any device. An open table there
  // has a `rejoin` path: a table in Now that this session is not bound
  // to gets Open or Resume, which trades this session for that seat's
  // (POST /me/games/{id}/session). The rest is History.
  let mine = $state<MyGame[]>([]);
  const mineByID = $derived(new Map(mine.map((g) => [g.id, g])));
  const rejoinable = $derived(new Set(mine.filter((g) => g.rejoin).map((g) => g.id)));
  let opening = $state<string | null>(null);

  async function loadMyGames(): Promise<void> {
    const userID = signedInUserID($session);
    if (!userID) {
      mine = [];
      return;
    }
    try {
      const list = await fetchMyGames();
      // The same read tells the tutorial offer whether this person has
      // finished a game (ADR 0125 §3.7, LobbyPractice.hint.ts).
      noteMyGames(userID, list);
      mine = list;
    } catch {
      // Keep what we had: a failed read must not hide a table.
    }
  }

  // rejoinsHere: a table where this person holds a seat that this
  // session is not bound to.
  function rejoinsHere(id: string): boolean {
    return rejoinable.has(id) && $session?.gameID !== id;
  }

  async function openMine(id: string, state: GameMeta["state"]): Promise<void> {
    const path = mineByID.get(id)?.rejoin;
    if (!path || opening) return;
    opening = id;
    error = "";
    try {
      await rejoinMyGame(path);
      // A table still in the lobby opens here, where the deck is
      // imported; one underway opens on the board.
      if (state === "lobby") await refresh();
      else navigate(`#/games/${id}`);
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "couldn't open that table";
    } finally {
      opening = null;
    }
  }

  let lastSetup = $state<TableSetup | null>(null);
  let useLastSetup = $state(true);
  // ADR 0159 §1: shown only when the server offers voice rooms.
  let wantVoiceRoom = $state(false);
  // The table this page just created: its tablemate picker starts open.
  let justCreated = $state<string | null>(null);
  // What applying the setup did, shown on the new table's card.
  let createdNote = $state<{ gameID: string; message: string } | null>(null);
  let seatBusy = $state<string | null>(null);

  // The invite tokens this page holds, per table. GET /games strips
  // both invites, so they come from the create response, a rotation,
  // or GET /games/{id} for the creator and a seated player. The
  // spectator invite is distinct from the player one: sharing the
  // player one with a spectator would let them claim a seat.
  const recentInvites = new SvelteMap<string, string>();
  const recentSpectatorInvites = new SvelteMap<string, string>();

  // --- bot seats (S31, ADR 0033 §9) -------------------------------
  //
  // Bots take real seats out of the same four, so the control lives on
  // the open seats. The option catalog is fetched once; `enabled` false
  // means this server has no bot host, and Add bot is not offered.
  let botOptions = $state<BotOptions | null>(null);
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
  // the server gate; the endpoint is authoritative.
  function canManageBots(g: GameMeta): boolean {
    if (g.state !== "lobby" || !botsOfferable) return false;
    if (isAdmin) return true;
    return mySeat(g) !== null;
  }

  // --- table settings (ADR 0075 §2.5) -----------------------------
  //
  // Offered to the host and the admin: the lobby's HTTP surface has
  // only the write, PATCH, so the read is an empty patch and inherits
  // the write's gate.
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
      tableSettings = await patchTableSettings(gameID, patch);
    } catch (e) {
      settingsError = e instanceof Error ? e.message : String(e);
    } finally {
      settingsBusy = false;
    }
  }

  // canRotateInvites: admin, or the table's own creator (#1098).
  // g.is_creator is computed server-side per viewer.
  function canRotateInvites(g: GameMeta): boolean {
    return isAdmin || g.is_creator === true;
  }

  function openBotPicker(gameID: string): void {
    botError = "";
    botPickerFor = gameID;
    botTier = defaultBotTier(botOptions?.tiers ?? []);
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
      botError = lobbyErrorMessage(err, "could not add the bot");
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

  // --- removing someone's seat (#2785) -----------------------------
  let confirmRemove = $state<{ gameID: string; seat: SeatInfo } | null>(null);
  let removeBusy = $state(false);
  // The confirm's "also block them" box (#2984), off each time it opens.
  let removeBlock = $state(false);
  let removeError = $state<{ gameID: string; message: string } | null>(null);
  // Bumped after a removal, so a block list on screen reads the new entry.
  let blockRefresh = $state(0);

  function canRemove(g: GameMeta, p: SeatInfo): boolean {
    return canRemoveSeat(g, p, { isAdmin, mySeat: mySeat(g) });
  }

  async function onRemoveSeat(gameID: string, playerID: string): Promise<void> {
    removeBusy = true;
    removeError = null;
    try {
      await removeSeat(gameID, playerID, { block: removeBlock });
      confirmRemove = null;
      blockRefresh += 1;
      await refresh();
    } catch (err) {
      removeError = {
        gameID,
        message: err instanceof LobbyApiError ? err.message : "could not remove the seat",
      };
    } finally {
      removeBusy = false;
    }
  }

  function botDeckName(id: string | undefined): string {
    if (!id) return "";
    return botOptions?.decks.find((d) => d.id === id)?.name ?? id;
  }

  const botTierNote = $derived(tierNote(botOptions?.tiers ?? [], botTier));
  const botDeckDescription = $derived(
    botOptions?.decks.find((d) => d.id === botDeck)?.description ?? "",
  );

  // --- archiving and deleting (admin) ------------------------------
  //
  // ARCHIVE hides a table and keeps everything; DELETE reaps the replay
  // and the restore point and cannot be undone, so it is offered only on
  // an archived table, behind its own confirmation. The admin is the
  // shared token, or an allowlisted person in admin mode (lib/admin.ts).
  const isAdmin = $derived(isAdminSession($session));
  let archived = $state<GameMeta[]>([]);
  let manageBusy = $state(false);
  let confirming = $state<{ id: string; kind: "archive" | "delete" } | null>(null);

  // Tables whose invites this page has already asked the server for.
  const askedInvites = new Set<string>();

  async function loadInvites(list: GameMeta[]): Promise<void> {
    for (const g of list) {
      // The creator (wherever they are) and a seated player at a table
      // still in the lobby: GET /games/{id} serves them the invites.
      const seated = $session?.gameID === g.id && g.state === "lobby";
      if (!g.is_creator && !seated) continue;
      if (recentInvites.has(g.id) || askedInvites.has(g.id)) continue;
      askedInvites.add(g.id);
      try {
        const full = await getGame(g.id);
        if (full.invite_token) recentInvites.set(g.id, full.invite_token);
        if (full.spectator_invite) recentSpectatorInvites.set(g.id, full.spectator_invite);
      } catch {
        // No cached link; Copy invite mints one for those who may.
      }
    }
  }

  async function refresh(): Promise<void> {
    try {
      const [list] = await Promise.all([listGames(), loadMyGames()]);
      await loadInvites(list);
      games = list;
      if (isAdmin) archived = await listGames({ archived: true });
      else archived = [];
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "list failed";
    } finally {
      loaded = true;
    }
  }

  // manage runs one admin table action with shared busy / error /
  // confirmation handling.
  async function manage(fn: () => Promise<unknown>, failure: string): Promise<void> {
    manageBusy = true;
    error = "";
    try {
      await fn();
      confirming = null;
      await refresh();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : failure;
    } finally {
      manageBusy = false;
    }
  }

  // --- admin: seat reclaim links ----------------------------------
  //
  // A reclaim link is a bearer credential for ONE player's seat, so it
  // is minted on demand, shown once, single use and short-lived, and
  // the panel says all three.
  let reclaim = $state<{ gameID: string; ticket: ReclaimTicket; url: string } | null>(null);
  let reclaimBusy = $state<string | null>(null);
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
    // A blank name is fine: the server picks one (#2630).
    busy = true;
    createError = "";
    try {
      const withSetup = useLastSetup && lastSetup !== null;
      const meta = await createGame(newName.trim(), {
        ...(withSetup ? { setup: "last" as const } : {}),
        voiceRoom: $community.voice && wantVoiceRoom,
      });
      if (meta.invite_token) recentInvites.set(meta.id, meta.invite_token);
      if (meta.spectator_invite) recentSpectatorInvites.set(meta.id, meta.spectator_invite);
      justCreated = meta.id;
      const note = setupResultMessage(meta.setup);
      createdNote = note ? { gameID: meta.id, message: note } : null;
      newName = "";
      newOpen = false;
      await refresh();
    } catch (err) {
      createError = lobbyErrorMessage(err, "create failed");
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

  // A transient "copied" flash keyed by "<gameID>:<kind>", so each
  // copy button flashes on its own. A failed write is surfaced:
  // writeText rejects on insecure origins and unfocused tabs.
  let copied = $state<string | null>(null);
  const COPY_FLASH_MS = 1500;
  async function copyToClipboard(url: string, key: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(url);
      copied = key;
      setTimeout(() => {
        if (copied === key) copied = null;
      }, COPY_FLASH_MS);
    } catch {
      error = "clipboard write failed — copy the link from a focused, HTTPS tab";
    }
  }

  // copyInvite copies the table's invite link (ADR 0158 §1 item 2). With
  // none held here, one who may rotate invites gets a new one minted
  // first. Nobody else ever sees the button without a link.
  async function copyInvite(g: GameMeta): Promise<void> {
    let token = recentInvites.get(g.id);
    if (!token) {
      if (!canRotateInvites(g)) return;
      rotateBusy = `${g.id}:player`;
      try {
        token = (await rotateInvite(g.id, "player")).token;
        recentInvites.set(g.id, token);
      } catch (err) {
        rotateError = {
          gameID: g.id,
          message: err instanceof LobbyApiError ? err.message : "could not make an invite link",
        };
        return;
      } finally {
        rotateBusy = null;
      }
    }
    await copyToClipboard(inviteURL(g.id, token), `${g.id}:invite`);
  }

  async function copySpectatorInvite(id: string): Promise<void> {
    const token = recentSpectatorInvites.get(id);
    if (!token) return;
    await copyToClipboard(spectatorInviteURL(id, token), `${id}:spectator`);
  }

  // --- rotating a lost or compromised invite ----------------------
  //
  // Rotating revokes the current invite of one kind and mints a
  // replacement. The old link stops working at once, so it asks first.
  type InviteKind = "player" | "spectator";
  let confirmRotate = $state<{ gameID: string; kind: InviteKind } | null>(null);
  let rotateBusy = $state<string | null>(null); // `${gameID}:${kind}`
  let rotateError = $state<{ gameID: string; message: string } | null>(null);
  let rotated = $state<{ gameID: string; kind: InviteKind; url: string } | null>(null);

  function inviteKindLabel(kind: InviteKind): string {
    return kind === "player" ? "player invite" : "spectator link";
  }

  async function onRotateInvite(gameID: string, kind: InviteKind): Promise<void> {
    rotateBusy = `${gameID}:${kind}`;
    rotateError = null;
    try {
      const res = await rotateInvite(gameID, kind);
      const url =
        kind === "player" ? inviteURL(gameID, res.token) : spectatorInviteURL(gameID, res.token);
      if (kind === "player") recentInvites.set(gameID, res.token);
      else recentSpectatorInvites.set(gameID, res.token);
      rotated = { gameID, kind, url };
      confirmRotate = null;
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

  // takeSeat sits the creator down at their own table, as themselves.
  // The new seat session replaces this page's session.
  async function takeSeat(g: GameMeta): Promise<void> {
    const invite = recentInvites.get(g.id);
    if (!invite) return;
    seatBusy = g.id;
    error = "";
    try {
      await joinGame(g.id, invite, "");
      await refresh();
    } catch (err) {
      error = lobbyErrorMessage(err, "could not take a seat");
    } finally {
      seatBusy = null;
    }
  }

  function openGame(id: string): void {
    navigate(`#/games/${id}`);
  }

  // mySeat returns the seat this SESSION occupies in game g, or null.
  function mySeat(g: GameMeta): SeatInfo | null {
    const s = $session;
    if (!s?.playerID || s.gameID !== g.id) return null;
    return g.players.find((p) => p.player_id === s.playerID) ?? null;
  }

  // mySeatNumber: the seat this person holds at g, through this session
  // or another one (GET /me/games).
  function mySeatNumber(g: GameMeta): number | null {
    return mySeat(g)?.seat ?? mineByID.get(g.id)?.seat ?? null;
  }

  // startsFromSeat: a seated host or creator, who may start the table.
  function startsFromSeat(g: GameMeta): boolean {
    const me = mySeat(g);
    return me !== null && (me.is_host === true || g.is_creator === true);
  }

  function canStart(g: GameMeta): boolean {
    return g.state === "lobby" && g.players.length >= 2 && g.players.every((p) => p.deck_uploaded);
  }

  // startReason explains why the table cannot start yet, in one line.
  function startReason(g: GameMeta): string {
    if (g.state !== "lobby") return "";
    const me = mySeat(g);
    if (me && !me.deck_uploaded) return "Choose your deck below";
    if (g.players.length < 2) return "Needs at least two seats";
    const pending = g.players.filter((p) => !p.deck_uploaded);
    if (pending.length === 1) return `Waiting on ${seatName(pending[0])}'s deck`;
    if (pending.length > 1) return `Waiting on ${pending.length} decks`;
    if (me && !startsFromSeat(g)) return "Waiting for the host to start";
    return "";
  }

  // The card's one primary button (ADR 0158 §1 item 2): Open in the
  // lobby, Resume in play. A seat held through another session is
  // reclaimed first. Whoever may start a table of two or more gets
  // Start in its place; Open moves to the ⋯ menu.
  interface Primary {
    label: string;
    run: () => void;
    disabled?: boolean;
    title?: string;
    busy?: boolean;
  }

  function mayStart(g: GameMeta): boolean {
    return g.state === "lobby" && g.players.length >= 2 && (startsFromSeat(g) || isAdmin);
  }

  function primaryFor(g: GameMeta): Primary {
    const held = mySeatNumber(g) !== null;
    if (rejoinsHere(g.id)) {
      return {
        label: g.state === "active" ? "Resume" : "Open",
        run: () => void openMine(g.id, g.state),
        disabled: opening !== null,
        busy: opening === g.id,
        title: "open your seat at this table",
      };
    }
    if (mayStart(g)) {
      const reason = startReason(g);
      return {
        label: "Start",
        run: () => void onStart(g.id),
        disabled: !canStart(g),
        title: reason || "start the game",
      };
    }
    if (g.state === "active") {
      return { label: held ? "Resume" : "Open", run: () => openGame(g.id) };
    }
    return { label: "Open", run: () => openGame(g.id) };
  }

  // Copy invite (§1 item 2): the host or creator, in the lobby, when
  // there is a link to copy or one they may make.
  function offersCopyInvite(g: GameMeta): boolean {
    if (g.state !== "lobby") return false;
    const mayShare = isAdmin || g.is_creator === true || mySeat(g)?.is_host === true;
    return mayShare && (recentInvites.has(g.id) || canRotateInvites(g));
  }

  function moreFor(g: GameMeta): MoreItem[] {
    const items: MoreItem[] = [];
    if (mayStart(g) && !rejoinsHere(g.id)) {
      items.push({ label: "Open table", onclick: () => openGame(g.id) });
    }
    if (g.state === "lobby" && g.is_creator && !mySeat(g) && recentInvites.has(g.id)) {
      items.push({
        label: seatBusy === g.id ? "Taking a seat…" : "Take a seat",
        disabled: seatBusy !== null,
        onclick: () => void takeSeat(g),
      });
    }
    if (recentSpectatorInvites.has(g.id)) {
      items.push({ label: "Copy spectator link", onclick: () => void copySpectatorInvite(g.id) });
    }
    if (canRotateInvites(g)) {
      items.push({
        label: "New invite link",
        title: "mint a new player invite — the current one stops working immediately",
        disabled: rotateBusy !== null,
        onclick: () => (confirmRotate = { gameID: g.id, kind: "player" }),
      });
      items.push({
        label: "New spectator link",
        title: "mint a new spectator link — the current one stops working immediately",
        disabled: rotateBusy !== null,
        onclick: () => (confirmRotate = { gameID: g.id, kind: "spectator" }),
      });
    }
    if (canManageTableFor(g)) {
      items.push({
        label: settingsFor === g.id ? "Hide table settings" : "Table settings",
        title: "undos, starting life, commander damage, table pace, spawning",
        onclick: () => (settingsFor === g.id ? (settingsFor = null) : openTableSettings(g.id)),
      });
    }
    // The server's replay gate: an admin any time after the lobby.
    if (isAdmin && g.state !== "lobby") {
      const url = replayURL(g.id);
      if (url) items.push({ label: "Download replay", href: url, download: `${g.id}.jsonl` });
    }
    if (isAdmin) {
      items.push({
        label: "Archive",
        title: "hide this table from the list — nothing is deleted",
        disabled: manageBusy,
        onclick: () => (confirming = { id: g.id, kind: "archive" }),
      });
    }
    return items;
  }

  // The card footer's two panels: invite a tablemate, and your deck.
  // Open on their own when they are the next thing to do (a table just
  // created; a deck not chosen yet), until the person toggles them.
  let panelChoice = $state<Record<string, "invite" | "deck" | null>>({});

  function panelFor(g: GameMeta): "invite" | "deck" | null {
    if (g.id in panelChoice) return panelChoice[g.id];
    if (justCreated === g.id) return "invite";
    const seat = mySeat(g);
    if (seat && g.state === "lobby" && !seat.deck_uploaded) return "deck";
    return null;
  }

  function togglePanel(g: GameMeta, which: "invite" | "deck"): void {
    panelChoice = { ...panelChoice, [g.id]: panelFor(g) === which ? null : which };
  }

  function canInvite(g: GameMeta): boolean {
    return (
      g.state === "lobby" &&
      canInviteTablemates($session, $session?.gameID, g.id, g.is_creator === true)
    );
  }

  function initials(p: SeatInfo): string {
    return seatName(p).trim().slice(0, 1).toUpperCase() || "?";
  }

  // A table in the lobby seats four, so its card pads with open seats
  // (where Add bot lives). A game in play shows who is there.
  const TABLE_SEATS = 4;
  function seatSlots(g: GameMeta): (SeatInfo | null)[] {
    const slots: (SeatInfo | null)[] = [...g.players].sort((a, b) => a.seat - b.seat);
    if (g.state === "lobby") while (slots.length < TABLE_SEATS) slots.push(null);
    return slots;
  }

  // What the page lists (lib/signedInHome.ts's myTables): every table
  // for an admin, a signed-in person's own, and a guest's view as it
  // always was. Now is the part of it in the lobby or in play.
  const visibleGames = $derived(myTables(games, $session, { admin: isAdmin, rejoinable }));
  const now = $derived(nowTables(visibleGames));
  const nowIDs = $derived(new Set(now.map((g) => g.id)));

  // --- History (§1 item 3) ------------------------------------------
  let historyFilter = $state<HistoryFilter>("all");
  let historyShown = $state(HISTORY_FIRST);
  let confirmDelete = $state<string | null>(null);

  const history = $derived(historyRows(mine, [...visibleGames, ...archived], nowIDs));
  const filtered = $derived(filterHistory(history, historyFilter));
  const shownRows = $derived(filtered.slice(0, historyShown));
  const more = $derived(moreCount(filtered.length, historyShown));

  function setFilter(f: HistoryFilter): void {
    historyFilter = f;
    historyShown = HISTORY_FIRST;
  }

  // Replay, as the server's gate allows it: an admin after the lobby,
  // and anyone who sat at a game once it has ended.
  function replayFor(r: HistoryRow): string | null {
    const allowed =
      (isAdmin && r.state !== "lobby") ||
      (r.state === "ended" && (r.seated || $session?.gameID === r.id));
    return allowed ? replayURL(r.id) : null;
  }

  function historyMore(r: HistoryRow): MoreItem[] {
    const items: MoreItem[] = [];
    if (rejoinsHere(r.id)) {
      items.push({
        label: "Open",
        disabled: opening !== null,
        onclick: () => void openMine(r.id, r.state),
      });
    } else if (isAdmin || $session?.gameID === r.id) {
      items.push({ label: "Open", onclick: () => openGame(r.id) });
    }
    if (isAdmin) {
      if (r.archived) {
        items.push({
          label: "Unarchive",
          disabled: manageBusy,
          onclick: () => void manage(() => unarchiveGame(r.id), "unarchive failed"),
        });
        items.push({
          label: "Delete…",
          danger: true,
          disabled: manageBusy,
          onclick: () => (confirmDelete = r.id),
        });
      } else {
        items.push({
          label: "Archive",
          disabled: manageBusy,
          onclick: () => void manage(() => archiveGame(r.id), "archive failed"),
        });
      }
    }
    return items;
  }

  // Poll while a table in Now can change (the lobby has no WebSocket),
  // and re-read on focus, so seats, decks, turns and life totals keep up.
  const POLL_MS = 5000;
  onMount(() => {
    void refresh();
    if (signedInUserID($session)) {
      void fetchMySetup()
        .then((r) => (lastSetup = r.setup ?? null))
        .catch(() => undefined);
    }
    void fetchBotOptions()
      .then((o) => (botOptions = o))
      .catch(() => undefined);
    const t = setInterval(() => {
      if (now.length > 0) void refresh();
    }, POLL_MS);
    const onFocus = (): void => void refresh();
    const onVisible = (): void => {
      if (document.visibilityState === "visible") void refresh();
    };
    window.addEventListener("focus", onFocus);
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      clearInterval(t);
      window.removeEventListener("focus", onFocus);
      document.removeEventListener("visibilitychange", onVisible);
    };
  });

  // #/lobby?show=history (and the old #/my-games, which redirects
  // there) scrolls to History once the lists have loaded.
  let scrolledToHistory = false;
  $effect(() => {
    const r = $route;
    if (!loaded || r.name !== "lobby" || r.show !== "history" || scrolledToHistory) return;
    scrolledToHistory = true;
    void tick().then(() => document.getElementById(HISTORY_ID)?.scrollIntoView?.());
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

{#snippet headActions()}
  {#if joinBox !== "none"}
    <form class="join" onsubmit={onJoin}>
      <input
        class="mono"
        type="text"
        placeholder={joinBox === "code" ? "invite code or link" : "invite link"}
        aria-label={joinBox === "code" ? "invite code or link" : "invite link"}
        title={joinBox === "code"
          ? "The code from your pod's invite, or the whole link. You sit as your Discord name; a spectator link opens the table read-only."
          : "The whole invite link. A spectator link opens the table read-only."}
        bind:value={joinInput}
      />
      <button type="submit" class="btn ghost" disabled={joining || !joinInput.trim()}>
        {joining ? "…" : "Join"}
      </button>
    </form>
  {/if}
  {#if canCreate}
    <button
      type="button"
      class="btn primary"
      aria-label={L.newTable}
      aria-expanded={newOpen}
      onclick={openNewTable}
    >
      + New table
    </button>
  {/if}
{/snippet}

<!-- The site header is the page's only chrome (ADR 0112 §1 item 4):
     Page draws it, and the title (ADR 0156 §3). -->
<Page
  eyebrow="Play"
  title="Tables"
  titleLabel={L.lobbyTitle}
  lede="Your tables now, and every game you've played."
  actions={joinBox !== "none" || canCreate ? headActions : undefined}
>
  <section class="lobby">
    <RemovedNotice />

    {#if modeNotice}
      <div class="mode-notice" role="status">
        <span>{modeNotice}</span>
        <button type="button" class="ghost" onclick={() => (modeNotice = "")}>dismiss</button>
      </div>
    {/if}

    {#if joinError}
      <p class="error" role="alert">{joinError}</p>
    {/if}

    {#if canCreate && newOpen}
      <form class="new-table" aria-label={L.createGame} onsubmit={onCreate}>
        <span class="nt-label">New table</span>
        <div class="nt-row"><NewTableRow bind:name={newName} {busy} /></div>
        {#if lastSetup}
          <label class="use-setup">
            <input type="checkbox" bind:checked={useLastSetup} />
            <span>
              Use my last setup
              <span class="setup-sum">{setupSummary(lastSetup, botDeckName)}</span>
            </span>
          </label>
        {/if}
        {#if $community.voice}
          <label class="use-setup">
            <input type="checkbox" bind:checked={wantVoiceRoom} />
            <span>
              Create a voice room
              <span class="setup-sum"
                >a private channel in the CMD & CTRL Discord for this table's players</span
              >
            </span>
          </label>
        {/if}
        {#if createError}
          <p class="error nt-error" role="alert">{createError}</p>
        {/if}
      </form>
    {/if}

    {#if error}
      <p class="error">{error}</p>
    {/if}

    <section class="now" aria-labelledby="now-h">
      <h2 class="sec-h" id="now-h">Now <span class="count">{now.length}</span></h2>

      {#if now.length === 0}
        {#if joinBox === "code"}
          <!-- ADR 0112 §1 item 7: a signed-in person with no tables. -->
          <div class="empty-card">
            <p>
              You're not at a table yet. Paste an invite code above, or create a table and invite
              your pod.
            </p>
          </div>
          <!-- ADR 0156 §6: three ways to start. The third is the C2
               Discord, only when the server names an invite. -->
          <div class="empty-cards">
            <div class="card">
              <h3>{L.practiceGame}</h3>
              <p>Five minutes against a bot that shows you where everything is.</p>
              <a class="btn ghost sm" href="#/practice">{L.startPractice}</a>
            </div>
            <div class="card">
              <h3>Bring a deck</h3>
              <p>Check how much of your list plays as printed, and save it.</p>
              <a class="btn ghost sm" href="#/decks">Decks</a>
            </div>
            {#if $community.invite_url}
              <div class="card pod">
                <h3>No pod yet?</h3>
                <p>Players post open seats on our Discord.</p>
                <a
                  class="btn discord sm"
                  href={$community.invite_url}
                  target="_blank"
                  rel="noopener noreferrer"><Icon name="discord" size={15} /> {L.findAPod}</a
                >
              </div>
            {/if}
          </div>
        {:else}
          <p class="muted empty">
            {#if canCreate}
              No tables right now. Make one with New table, then send the invite link to your pod.
            {:else}
              You're not seated at a table yet. Ask for an invite link.
            {/if}
          </p>
        {/if}
      {:else}
        <ul class="games" aria-label="tables now">
          {#each now as g (g.id)}
            {@const seat = mySeat(g)}
            {@const myNo = mySeatNumber(g)}
            {@const chip = statusChip(g)}
            {@const primary = primaryFor(g)}
            {@const items = moreFor(g)}
            {@const reason = startReason(g)}
            {@const panel = panelFor(g)}
            <li class="tcard" class:mine={myNo !== null}>
              <div class="thead">
                <div class="tid">
                  <h3 class="tname">{g.name}</h3>
                  <span class="chip" class:live={chip.live}>
                    {#if chip.live}<i class="dot ok"></i>{/if}{chip.label}
                  </span>
                  <span class="ctx">{contextLine(g, { seat: myNo })}</span>
                </div>
                <div class="tact">
                  {#if offersCopyInvite(g)}
                    <button
                      type="button"
                      class="btn ghost sm"
                      class:on={copied === `${g.id}:invite`}
                      disabled={rotateBusy !== null}
                      title={recentInvites.has(g.id)
                        ? "copy this table's invite link"
                        : "no link is held here: this makes a new one, and an older link stops working"}
                      onclick={() => copyInvite(g)}
                    >
                      {#if copied === `${g.id}:invite`}
                        <Icon name="check" size={13} /> Invite copied
                      {:else}
                        <Icon name="link" size={13} /> Copy invite
                      {/if}
                    </button>
                  {/if}
                  {#if items.length > 0}
                    <MoreMenu label={`more for ${g.name}`} {items} />
                  {/if}
                  <button
                    type="button"
                    class="btn primary sm"
                    disabled={primary.disabled}
                    title={primary.title}
                    onclick={primary.run}
                  >
                    {primary.busy ? "…" : primary.label}
                  </button>
                </div>
              </div>

              <ul class="seats" aria-label="seats">
                {#each seatSlots(g) as p, i (p?.player_id ?? `open-${i}`)}
                  {#if p}
                    {@const avatar = avatarURL(p.discord_id, p.discord_avatar_hash)}
                    {@const line = seatLine(g, p, botDeckName)}
                    <li class="seat" class:you={myNo === p.seat} class:bot={p.is_bot}>
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
                          {seatName(p)}
                          {#if p.is_host}<b
                              title="Table host: may manage the table alongside the admin">host</b
                            >{/if}
                          {#if myNo === p.seat}<b>you</b>{/if}
                        </div>
                        {#if p.is_agent && !p.is_bot}
                          <div class="sagent"><AgentChip seat={p} /></div>
                        {/if}
                        <div class="sstat" class:pend={line.pending}>
                          {#if !line.pending && (p.is_bot || p.deck_uploaded)}<i class="dot ok"
                            ></i>{/if}{line.text}
                        </div>
                      </div>
                      {#if !p.is_bot && isAdmin}
                        <!-- Admin-only: the disconnected player's way back in. -->
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
                      {#if !p.is_bot && canRemove(g, p)}
                        <button
                          class="seat-link seat-remove"
                          title={`remove ${seatName(p)} from this table`}
                          aria-label={`remove ${seatName(p)}`}
                          disabled={removeBusy}
                          onclick={() => {
                            removeBlock = false;
                            confirmRemove = { gameID: g.id, seat: p };
                          }}
                        >
                          remove
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
                        <div class="sstat dim">send the invite link</div>
                      </div>
                      {#if canManageBots(g) && botPickerFor !== g.id}
                        <button
                          class="seat-add"
                          disabled={botBusy}
                          onclick={() => openBotPicker(g.id)}
                        >
                          <Icon name="robot" size={13} /> Add bot
                        </button>
                      {/if}
                    </li>
                  {/if}
                {/each}
              </ul>

              {#if botPickerFor === g.id}
                <div class="bot-picker">
                  <label>
                    <span>tier</span>
                    <select bind:value={botTier} disabled={botBusy}>
                      {#each botOptions?.tiers ?? [] as t (t.tier)}
                        <option value={t.tier} disabled={!t.available} title={t.description}>
                          {tierOptionLabel(t)}
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
                  {#if botTierNote}
                    <p class="deck-desc" class:tier-warning={botTierNote.warning}>
                      {botTierNote.text}
                    </p>
                  {/if}
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
                    A bot takes a real seat, arrives with its deck already loaded, and starts
                    playing when you press Start.
                  </p>
                </div>
              {:else if botError}
                <div class="bot-error">{botError}</div>
              {/if}

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

              {#if confirming?.id === g.id && confirming.kind === "archive"}
                <div class="confirm" role="alert">
                  <p>
                    <strong>Archive “{g.name}”?</strong>
                    It leaves Now and stops taking part — bots stop, and anyone still connected is dropped.
                    Nothing is deleted: the board, the replay and the invite links are all kept, and you
                    can unarchive it from History.
                  </p>
                  <div class="confirm-actions">
                    <button
                      class="primary"
                      disabled={manageBusy}
                      onclick={() => manage(() => archiveGame(g.id), "archive failed")}
                    >
                      archive table
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
                    The current {inviteKindLabel(confirmRotate.kind)} link stops working the moment this
                    happens — anyone still holding it will need the new one.
                  </p>
                  <div class="confirm-actions">
                    <button
                      class="primary"
                      disabled={rotateBusy !== null}
                      onclick={() => onRotateInvite(g.id, confirmRotate!.kind)}
                    >
                      {rotateBusy === `${g.id}:${confirmRotate.kind}`
                        ? "minting…"
                        : "mint new link"}
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
                    Send this to {reclaim.ticket.player_name} and nobody else. Anyone who opens it plays
                    as them — their hand included. It stops working the moment it is used.
                  </p>
                </div>
              {/if}
              {#if reclaimError?.gameID === g.id}
                <div class="bot-error">{reclaimError.message}</div>
              {/if}

              {#if confirmRemove?.gameID === g.id}
                <div class="confirm danger" role="alert">
                  <p>
                    <strong>{removeSeatQuestion(seatName(confirmRemove.seat))}</strong>
                    {removeSeatConsequence(g)}
                  </p>
                  <label class="block-opt">
                    <input type="checkbox" bind:checked={removeBlock} disabled={removeBusy} />
                    <span>{BLOCK_LABEL}</span>
                  </label>
                  <p class="hint">{BLOCK_HINT}</p>
                  <div class="confirm-actions">
                    <button
                      class="primary"
                      disabled={removeBusy}
                      onclick={() => onRemoveSeat(g.id, confirmRemove!.seat.player_id)}
                    >
                      {removeBusy ? "removing…" : "remove from table"}
                    </button>
                    <button
                      class="ghost"
                      disabled={removeBusy}
                      onclick={() => (confirmRemove = null)}
                    >
                      cancel
                    </button>
                  </div>
                </div>
              {/if}
              {#if removeError?.gameID === g.id}
                <div class="bot-error">{removeError.message}</div>
              {/if}

              <!-- #3113: who this table has blocked, with an Unblock. -->
              {#if canManageBlocks(g, { isAdmin, isHost: mySeat(g)?.is_host === true })}
                <BlockList gameID={g.id} refreshKey={blockRefresh} />
              {/if}

              {#if createdNote?.gameID === g.id}
                <p class="hint setup-note" role="status">{createdNote.message}</p>
              {/if}

              {#if g.state === "lobby" && (canInvite(g) || (seat && $session?.playerID) || reason)}
                <!-- The lobby's own controls (§1 item 2): invite a
                     tablemate, your deck, and what the table waits on. -->
                <div class="tfoot">
                  {#if canInvite(g)}
                    <button
                      type="button"
                      class="foot-toggle"
                      aria-expanded={panel === "invite"}
                      onclick={() => togglePanel(g, "invite")}
                    >
                      Invite a tablemate <i class="caret" class:up={panel === "invite"}></i>
                    </button>
                  {/if}
                  {#if seat && $session?.playerID}
                    <button
                      type="button"
                      class="foot-toggle"
                      class:pend={!seat.deck_uploaded}
                      aria-expanded={panel === "deck"}
                      onclick={() => togglePanel(g, "deck")}
                    >
                      {seat.deck_uploaded
                        ? `Change deck · ${seat.deck_name || "deck ready"}`
                        : "Choose your deck"}
                      <i class="caret" class:up={panel === "deck"}></i>
                    </button>
                  {/if}
                  {#if reason}<span class="reason">{reason}</span>{/if}
                </div>
                {#if panel === "invite" && canInvite(g)}
                  <div class="invite-picker">
                    <TablematePicker gameID={g.id} lastTable={lastSetup?.tablemates ?? []} />
                  </div>
                {:else if panel === "deck" && seat && $session?.playerID}
                  <div class="deck-upload">
                    <p class="hint">
                      Every deck is validated against Commander rules — 100-card singleton, color
                      identity, format legality — whichever way it arrives.
                    </p>
                    <DeckUploadForm
                      gameID={g.id}
                      playerID={$session.playerID}
                      onSuccess={() => void refresh()}
                    />
                  </div>
                {/if}
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </section>

    <section class="history" id={HISTORY_ID} aria-labelledby="history-h">
      <div class="sec-head">
        <h2 class="sec-h" id="history-h">
          History <span class="count">{history.length} game{history.length === 1 ? "" : "s"}</span>
        </h2>
        {#if history.length > 0}
          <div class="seg" role="group" aria-label="filter history">
            {#each HISTORY_FILTERS as f (f.id)}
              <button
                type="button"
                aria-pressed={historyFilter === f.id}
                class:on={historyFilter === f.id}
                onclick={() => setFilter(f.id)}>{f.label}</button
              >
            {/each}
          </div>
        {/if}
      </div>

      {#if history.length === 0}
        <p class="muted empty">
          {#if signedInUserID($session) || isAdmin}
            Games you finish show up here, with their replays.
          {:else}
            Your games are kept for your Discord account. Sign in with Discord to see them here.
          {/if}
        </p>
      {:else}
        <div class="hist">
          <div class="hrow hhead" aria-hidden="true">
            <span>Table</span><span>Result</span><span>With</span><span>When</span><span></span>
          </div>
          {#if filtered.length === 0}
            <p class="muted hist-empty">No games match this filter.</p>
          {/if}
          <ul class="hlist" aria-label="history">
            {#each shownRows as r (r.id)}
              {@const replay = replayFor(r)}
              {@const ritems = historyMore(r)}
              <li class="hrow">
                <span class="hname">{r.name}</span>
                <span><span class="chip result {r.result.tone}">{r.result.label}</span></span>
                <span class="hwith">{r.others || "nobody else"}</span>
                <span class="hwhen">{whenLabel(r)}</span>
                <span class="hact">
                  {#if replay}
                    <a class="btn ghost sm" href={replay} download={`${r.id}.jsonl`}>
                      <Icon name="draw" size={13} /> Replay
                    </a>
                  {/if}
                  {#if ritems.length > 0}
                    <MoreMenu label={`more for ${r.name}`} items={ritems} />
                  {/if}
                </span>
                {#if confirmDelete === r.id}
                  <div class="confirm danger hconfirm" role="alert">
                    <p>
                      <strong>Delete “{r.name}” permanently?</strong>
                      This destroys the board, the restore point and the replay file. It cannot be undone.
                    </p>
                    <div class="confirm-actions">
                      <button
                        class="primary"
                        disabled={manageBusy}
                        onclick={() =>
                          manage(async () => {
                            await deleteGame(r.id);
                            confirmDelete = null;
                          }, "delete failed")}
                      >
                        delete permanently
                      </button>
                      <button
                        class="ghost"
                        disabled={manageBusy}
                        onclick={() => (confirmDelete = null)}
                      >
                        cancel
                      </button>
                    </div>
                  </div>
                {/if}
              </li>
            {/each}
          </ul>
          {#if more > 0}
            <button type="button" class="show-more" onclick={() => (historyShown += HISTORY_STEP)}>
              Show {more} more
            </button>
          {/if}
        </div>
      {/if}
    </section>
  </section>
</Page>

<style>
  .lobby {
    display: flex;
    flex-direction: column;
    gap: 18px;
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

  /* --- the page head: join and + New table (§1 item 1) --- */
  .join {
    display: flex;
    gap: 8px;
  }
  .join input {
    width: 230px;
    height: 40px;
    margin: 0;
    padding: 0 12px;
    box-sizing: border-box;
    font-family: var(--font-mono);
    font-size: 12.5px;
  }
  .new-table {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px 14px;
    padding: 12px 14px;
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
  }
  .nt-label {
    font-family: var(--font-mono);
    font-size: 10.5px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .nt-row {
    flex: 1 1 360px;
    min-width: 0;
  }
  .nt-error {
    flex: 1 0 100%;
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
    margin: 0;
  }
  .mode-notice {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
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
    margin: 0;
    font-size: 13.5px;
  }

  /* --- the two sections --- */
  .now,
  .history {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .history {
    margin-top: 18px;
    scroll-margin-top: 16px;
  }
  .sec-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }
  .sec-h {
    margin: 0;
    font-family: var(--font-display);
    font-size: 21px;
    font-weight: 700;
    color: var(--fg);
    text-transform: none;
    letter-spacing: -0.01em;
    display: flex;
    align-items: baseline;
    gap: 10px;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 500;
    color: var(--fg-dim);
  }

  /* A signed-in person's empty Now (ADR 0112 §1 item 7). */
  .empty-card {
    padding: 18px 20px;
    border-radius: 12px;
    border: 1px dashed var(--border-strong);
  }
  .empty-card p {
    margin: 0;
    font-size: 13.5px;
    color: var(--fg);
  }
  .empty-cards {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px;
  }
  .empty-cards .card {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }
  .empty-cards .card.pod {
    border-color: color-mix(in srgb, #5865f2 45%, var(--border));
    background: color-mix(in srgb, #5865f2 7%, var(--surface));
  }
  .empty-cards h3 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 700;
    color: var(--fg);
  }
  .empty-cards p {
    flex: 1;
    margin: 6px 0 16px;
    font-size: 13.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  @media (max-width: 819px) {
    .empty-cards {
      grid-template-columns: minmax(0, 1fr);
    }
  }

  /* --- Now: one card per table --- */
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
    padding: 18px 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .tcard.mine {
    border-color: color-mix(in srgb, var(--accent) 45%, transparent);
  }
  .thead {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px 20px;
    flex-wrap: wrap;
  }
  .tid {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px 12px;
    min-width: 0;
  }
  .tname {
    margin: 0;
    font-family: var(--font-display);
    font-size: 18px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--fg);
  }
  .ctx {
    font-size: 12.5px;
    color: var(--fg-dim);
  }
  .tact {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-left: auto;
  }
  .tact .btn.on {
    color: var(--mint);
    border-color: color-mix(in srgb, var(--mint) 40%, transparent);
  }
  .btn:disabled {
    opacity: 0.55;
    cursor: default;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 22px;
    padding: 0 9px;
    border-radius: 999px;
    border: 1px solid var(--accent-line);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--accent-strong);
    font-weight: 600;
    white-space: nowrap;
  }
  .chip.live {
    color: var(--mint);
    border-color: color-mix(in srgb, var(--mint) 45%, transparent);
    background: color-mix(in srgb, var(--mint) 8%, transparent);
  }
  .dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    flex: 0 0 auto;
  }
  .dot.ok {
    background: var(--mint);
  }

  .seats {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px;
  }
  /* Four seats in one row from the laptop up; two by two on a tablet,
     where a fourth of the card is too narrow for a name and Add bot. */
  @media (max-width: 1023px) {
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
    border: 1px solid var(--border);
    min-width: 0;
  }
  .seat.you {
    border-color: color-mix(in srgb, var(--accent) 40%, transparent);
  }
  .seat.open {
    border: 1px dashed var(--border-strong);
    background: transparent;
  }
  .sav {
    width: 38px;
    height: 38px;
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
    flex: 1 1 auto;
  }
  .sname {
    font-size: 13px;
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
    color: var(--accent-strong);
    font-weight: 700;
    margin-left: 5px;
  }
  .sname.dim,
  .sstat.dim {
    color: var(--fg-dim);
    font-weight: 500;
  }
  .sstat {
    font-size: 11.5px;
    color: var(--fg-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-top: 2px;
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .sstat.pend {
    color: var(--accent-strong);
  }
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
  .sagent {
    margin: 2px 0;
    min-width: 0;
  }
  .seat-x,
  .seat-add,
  .seat-link {
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
  .seat-add {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    height: 30px;
    padding: 0 10px;
    font-size: 12px;
    font-weight: 700;
    color: var(--fg);
    white-space: nowrap;
    background: var(--surface-raised);
  }
  .seat-link {
    font-size: 11px;
    padding: 4px 8px;
    white-space: nowrap;
  }
  .seat-x:hover:not(:disabled),
  .seat-add:hover:not(:disabled),
  .seat-link:hover:not(:disabled) {
    color: var(--fg);
    border-color: var(--accent);
  }
  .seat-x:disabled,
  .seat-add:disabled,
  .seat-link:disabled {
    opacity: 0.5;
    cursor: default;
  }

  /* --- the card footer: the lobby's own controls --- */
  .tfoot {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 20px;
    padding-top: 12px;
    border-top: 1px solid var(--border);
  }
  .foot-toggle {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: auto;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--fg-muted);
    font: 600 12.5px/1.4 var(--font-ui);
    cursor: pointer;
  }
  .foot-toggle:hover {
    color: var(--fg);
  }
  .foot-toggle.pend {
    color: var(--accent-strong);
  }
  .caret {
    width: 5px;
    height: 5px;
    border-right: 1.5px solid currentColor;
    border-bottom: 1.5px solid currentColor;
    transform: translateY(-2px) rotate(45deg);
    transition: transform 120ms var(--ease);
  }
  .caret.up {
    transform: translateY(1px) rotate(-135deg);
  }
  .reason {
    margin-left: auto;
    font-size: 12px;
    color: var(--fg-dim);
  }
  .deck-upload,
  .invite-picker {
    padding: 12px 14px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface-sunken);
  }
  .deck-upload .hint {
    margin-top: 0;
  }

  /* --- the card's inset panels: settings, confirmations, new links --- */
  .tsettings,
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
  .confirm .block-opt {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
    font-size: 12.5px;
    color: var(--fg);
  }
  .confirm .block-opt input {
    margin: 0;
  }
  .confirm .hint {
    margin-top: 4px;
    font-size: 11.5px;
    color: var(--fg-dim);
  }
  .confirm-actions {
    display: flex;
    gap: 8px;
    margin-top: 10px;
  }
  .reclaim {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px 14px;
    border: 1px solid color-mix(in srgb, var(--accent) 35%, transparent);
    border-radius: 10px;
    background: var(--accent-soft);
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
    border-color: color-mix(in srgb, var(--mint) 40%, transparent);
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
  .deck-desc {
    flex: 1 0 100%;
    margin: 0;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .deck-desc.tier-warning {
    color: var(--danger);
  }
  .bot-error {
    color: var(--danger);
    font-size: 12px;
  }
  .hint {
    font-size: 12px;
    color: var(--fg-muted);
    line-height: 1.5;
    margin: 10px 0 0;
  }

  /* --- History: one compact row per game (§1 item 3) --- */
  .seg {
    display: inline-flex;
    padding: 3px;
    gap: 2px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface);
  }
  .seg button {
    height: 28px;
    padding: 0 12px;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--fg-muted);
    font: 600 13px/1 var(--font-ui);
    cursor: pointer;
  }
  .seg button:hover {
    color: var(--fg);
  }
  .seg button.on {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .hist {
    border: 1px solid var(--border);
    border-radius: 14px;
    background: var(--surface);
  }
  .hlist {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .hrow {
    display: grid;
    /* Fixed result, when and action columns, so the header row (outside
       the list) and every game row line up whatever each row holds. */
    grid-template-columns: minmax(0, 2.4fr) 132px minmax(0, 1.6fr) 104px 148px;
    align-items: center;
    gap: 12px 16px;
    padding: 12px 18px;
    border-top: 1px solid var(--border);
  }
  .hhead {
    border-top: 0;
    padding-top: 12px;
    padding-bottom: 10px;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .hname {
    font-weight: 700;
    font-size: 14px;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .hwith,
  .hwhen {
    font-size: 13px;
    color: var(--fg-muted);
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .hwhen {
    color: var(--fg-dim);
  }
  .hact {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
  }
  .chip.result.won {
    color: var(--accent-strong);
    border-color: var(--accent-line);
    background: var(--accent-soft);
  }
  .chip.result.ended,
  .chip.result.open {
    color: var(--fg-muted);
    border-color: var(--border-strong);
  }
  .chip.result.archived {
    color: var(--fg-dim);
    border-color: var(--border-strong);
    border-style: dashed;
  }
  .hconfirm {
    grid-column: 1 / -1;
  }
  .hist-empty {
    margin: 0;
    padding: 14px 18px;
    border-top: 1px solid var(--border);
    font-size: 13px;
  }
  .show-more {
    display: block;
    width: 100%;
    height: 44px;
    border: 0;
    border-top: 1px solid var(--border);
    border-radius: 0 0 14px 14px;
    background: transparent;
    color: var(--fg-muted);
    font: 600 13.5px/1 var(--font-ui);
    cursor: pointer;
  }
  .show-more:hover {
    color: var(--fg);
    background: var(--overlay-faint);
  }
</style>
