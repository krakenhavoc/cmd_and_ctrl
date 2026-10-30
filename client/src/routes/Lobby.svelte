<script lang="ts">
  import { onMount, tick } from "svelte";
  import { SvelteMap } from "svelte/reactivity";
  import {
    archiveGame,
    createGame,
    deleteGame,
    fetchBotOptions,
    fetchPrebuiltDecks,
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
    type BotDeckInfo,
    type BotOptions,
    type GameMeta,
    type ReclaimTicket,
    type SeatInfo,
    type UploadDeckResponse,
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
    botChipText,
    contextLine,
    filterHistory,
    HISTORY_FILTERS,
    HISTORY_FIRST,
    HISTORY_ID,
    HISTORY_STEP,
    historyRows,
    lobbyPrimary,
    lobbySlots,
    moreCount,
    nowTables,
    readiness,
    seatLine,
    seatName,
    seatPreview,
    seatTileText,
    statusChip,
    whenLabel,
    type HistoryFilter,
    type HistoryRow,
    type LobbyPrimaryKind,
  } from "../lib/tablesPage";
  import type { TableSettingsView } from "../lib/protocol";
  import { settingsSummary, type TableSettingsPatch } from "../lib/tableSettings";
  import DeckPicker from "../lib/components/DeckPicker.svelte";
  import BotPicker from "../lib/components/BotPicker.svelte";
  import type { PickerCommit, PickerState } from "../lib/deckPicker";
  import GameCardsPanel from "../lib/components/GameCardsPanel.svelte";
  import { visibleShops } from "../lib/shopLinksStore";
  import type { PrebuiltDeck } from "../lib/prebuiltDecks";
  import TableSettingsPanel from "../lib/components/TableSettingsPanel.svelte";
  import TablematePicker from "../lib/components/TablematePicker.svelte";
  import LfgPost from "../lib/components/LfgPost.svelte";
  import { LFG_GUEST_HINT, lfgHeader, setLFG } from "../lib/lfgPost";
  import ResultsPost from "../lib/components/ResultsPost.svelte";
  import { setResults } from "../lib/resultsPost";
  import Icon from "../lib/components/Icon.svelte";
  import MoreMenu, { type MoreItem } from "../lib/components/MoreMenu.svelte";
  import ContextMenu from "../lib/components/ContextMenu.svelte";
  import SeatTile from "../lib/components/SeatTile.svelte";
  import { contextMenuTrigger, type MenuPoint } from "../lib/menuTrigger";
  import RemovedNotice from "../lib/components/RemovedNotice.svelte";
  import {
    LEAVE_SEAT_QUESTION,
    REMOVE_BOT_CONSEQUENCE,
    canRemoveSeat,
    isOwnSeat,
    leaveSeatConsequence,
    removeSeatConsequence,
    removeSeatQuestion,
  } from "../lib/seatRemoval";
  import { BLOCK_HINT, BLOCK_LABEL, canManageBlocks } from "../lib/moderation";
  import BlockList from "../lib/components/BlockList.svelte";
  import AgentChip from "../lib/components/AgentChip.svelte";
  import DeckPowerBadge from "../lib/components/DeckPowerBadge.svelte";
  import Page from "../lib/components/Page.svelte";
  import { noteMyGames } from "../lib/hints/endedGame";

  // Tables, the signed-in home (ADR 0112 §1, ADR 0158 §1). The page
  // head holds the join box and + New table. Under it, two sections:
  //
  //   - Now: the tables in the lobby or in play that this person sits
  //     at or created (every table for an admin), one card each. A
  //     table in the lobby is a character select (ADR 0212): the
  //     settings summary in its head, four seat tiles, the deck picker
  //     (or the bot picker) under the seats, and a footer with the
  //     readiness line, Open table, the Invite menu and one primary that
  //     follows the table's state. A table in play keeps its primary in
  //     the head (Resume or Open), its seats with life totals and a rules
  //     line. A right-click on either opens every action, and the
  //     admin's, as a menu; there is no ⋯ on a card.
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
  // ADR 0161 §1: shown only when the server has an #lfg board, off by
  // default. The post names its host, so it needs a Discord sign-in: the
  // shared admin token has none, and gets the box disabled.
  let wantLFG = $state(false);
  const lfgCreateBlocked = $derived(signedInUserID($session) === null);
  // ADR 0174 §1: shown only when the server has a results channel, off
  // by default. The post names the seats, not accounts, so a guest host
  // may tick it.
  let wantResults = $state(false);
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
  // the open seats: Add bot opens the bot picker under the seats (ADR
  // 0212 owner decision 8, BotPicker.svelte), and the footer's primary
  // adds it. The option catalog is fetched once; `enabled` false means
  // this server has no bot host, and Add bot is not offered.
  let botOptions = $state<BotOptions | null>(null);
  let botPickerFor = $state<string | null>(null);
  // What the open bot picker would add, for the footer's primary.
  let botCommit = $state<PickerCommit | null>(null);
  let botPicker = $state<{ commit: () => Promise<void> } | undefined>();
  const botBusy = $derived(botCommit?.busy === true);

  const botTiersAvailable = $derived((botOptions?.tiers ?? []).filter((t) => t.available));
  const botsOfferable = $derived(
    botOptions?.enabled === true &&
      (botTiersAvailable.length > 0 || botOptions?.own_key !== undefined) &&
      (botOptions?.decks.length ?? 0) > 0,
  );

  // canManageBots: admin, or a player seated at this table. Mirrors
  // the server gate; the endpoint is authoritative.
  function canManageBots(g: GameMeta): boolean {
    if (g.state !== "lobby" || !botsOfferable) return false;
    if (isAdmin) return true;
    return mySeat(g) !== null;
  }

  // --- table settings (ADR 0075 §2.5, #3499) -----------------------
  //
  // The card's rules line shows everyone at the table what differs from
  // the defaults, read with GET /games/{id}/settings: a seat, a
  // spectator, the creator or the admin. Clicking it opens the panel,
  // which only the host and the admin can change.
  let settingsFor = $state<string | null>(null);
  let tableSettings = $state<TableSettingsView | null>(null);
  let settingsBusy = $state(false);
  let settingsError = $state<string | null>(null);
  const cardSettings = new SvelteMap<string, TableSettingsView>();

  function canManageTableFor(g: GameMeta): boolean {
    return isAdmin || mySeat(g)?.is_host === true;
  }

  // canReadSettings mirrors the server's gate for the read.
  function canReadSettings(g: GameMeta): boolean {
    return isAdmin || $session?.gameID === g.id || g.is_creator === true;
  }

  // loadCardSettings fills the rules lines. A table still in the lobby is
  // re-read on every refresh, since that is when its host is changing
  // them; one in play is read once (the board is where it changes).
  async function loadCardSettings(list: GameMeta[]): Promise<void> {
    const want = list.filter(
      (g) =>
        g.state !== "ended" &&
        canReadSettings(g) &&
        (g.state === "lobby" || !cardSettings.has(g.id)),
    );
    await Promise.all(
      want.map(async (g) => {
        try {
          cardSettings.set(g.id, await fetchTableSettings(g.id));
        } catch {
          // No line beats a wrong one; the panel says why when opened.
        }
      }),
    );
  }

  async function openTableSettings(gameID: string): Promise<void> {
    settingsFor = gameID;
    tableSettings = cardSettings.get(gameID) ?? null;
    settingsError = null;
    settingsBusy = true;
    try {
      tableSettings = await fetchTableSettings(gameID);
      cardSettings.set(gameID, tableSettings);
    } catch (e) {
      settingsError = e instanceof Error ? e.message : String(e);
    } finally {
      settingsBusy = false;
    }
  }

  function toggleTableSettings(gameID: string): void {
    if (settingsFor === gameID) settingsFor = null;
    else void openTableSettings(gameID);
  }

  // The table's #lfg post (ADR 0161). The answer is the table's meta;
  // the list's next poll carries the post's status as the server
  // reports it.
  let lfgBusy = $state(false);
  let lfgError = $state<string | null>(null);
  // The table the last #lfg error belongs to, so the card shows it too.
  let lfgErrorFor = $state<string | null>(null);
  async function toggleLFG(gameID: string, enabled: boolean): Promise<void> {
    lfgBusy = true;
    lfgError = null;
    lfgErrorFor = null;
    try {
      await setLFG(gameID, enabled);
      await refresh();
    } catch (e) {
      lfgError = lobbyErrorMessage(e, "the #looking-for-game request failed");
      lfgErrorFor = gameID;
    } finally {
      lfgBusy = false;
    }
  }

  // The table's results post (ADR 0174). The answer is the table's meta;
  // the list's next poll carries it.
  let resultsBusy = $state(false);
  let resultsError = $state<string | null>(null);
  async function toggleResults(gameID: string, enabled: boolean): Promise<void> {
    resultsBusy = true;
    resultsError = null;
    try {
      await setResults(gameID, enabled);
      await refresh();
    } catch (e) {
      resultsError = lobbyErrorMessage(e, "the results request failed");
    } finally {
      resultsBusy = false;
    }
  }

  async function applyTableSettings(gameID: string, patch: TableSettingsPatch): Promise<void> {
    settingsBusy = true;
    settingsError = null;
    try {
      tableSettings = await patchTableSettings(gameID, patch);
      cardSettings.set(gameID, tableSettings);
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

  // openBotPicker is an open seat's Add bot. The bot picker takes the
  // region under the seats, so the deck picker or the tablemate picker
  // there closes, and focus moves into it.
  async function openBotPicker(g: GameMeta): Promise<void> {
    botPickerFor = g.id;
    panelChoice = { ...panelChoice, [g.id]: null };
    await tick();
    const region = document.getElementById(botPickerID(g.id));
    region?.scrollIntoView?.({ block: "nearest" });
    region?.focus({ preventScroll: true });
  }

  function botPickerID(gameID: string): string {
    return `bot-picker-${gameID}`;
  }

  // The seat a bot takes: the first open one, as the server seats it.
  function botSeatFor(g: GameMeta): number {
    const open = lobbySlots(g).findIndex((p) => p === null);
    return open >= 0 ? open : g.players.length;
  }

  // --- removing a seat (#2785, #3494) -----------------------------
  //
  // Every seat's × asks first: a person's, a bot's, and your own.
  // `kind` picks the route and the words: a bot goes through the bot
  // route; "self" is leaving your own seat, the host's included.
  type RemoveKind = "person" | "bot" | "self";
  let confirmRemove = $state<{ gameID: string; seat: SeatInfo; kind: RemoveKind } | null>(null);
  let removeBusy = $state(false);
  // The confirm's "also block them" box (#2984), off each time it opens.
  let removeBlock = $state(false);
  let removeError = $state<{ gameID: string; message: string } | null>(null);
  // Bumped after a removal, so a block list on screen reads the new entry.
  let blockRefresh = $state(0);

  function canRemove(g: GameMeta, p: SeatInfo): boolean {
    return canRemoveSeat(g, p, { isAdmin, mySeat: mySeat(g) });
  }

  function askRemove(g: GameMeta, p: SeatInfo): void {
    removeBlock = false;
    const kind: RemoveKind = p.is_bot
      ? "bot"
      : isOwnSeat(p, { isAdmin, mySeat: mySeat(g) })
        ? "self"
        : "person";
    confirmRemove = { gameID: g.id, seat: p, kind };
  }

  async function onRemoveSeat(gameID: string, playerID: string, kind: RemoveKind): Promise<void> {
    removeBusy = true;
    removeError = null;
    try {
      if (kind === "bot") await removeBotSeat(gameID, playerID);
      else await removeSeat(gameID, playerID, { block: kind === "person" && removeBlock });
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

  // The bot decks are the pre-built decks, so GET /decks holds each
  // one's coverage for the bot picker's tiles and its list for "Play
  // this deck in paper" (ADR 0179 §7).
  let prebuiltDecks = $state<PrebuiltDeck[]>([]);

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
      void loadCardSettings(list);
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
        lfg: $community.lfg === true && !lfgCreateBlocked && wantLFG,
        results: $community.results === true && wantResults,
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
  // copyToClipboard answers whether the write landed. `quiet` leaves the
  // failure to the caller, which has somewhere better to put the link.
  async function copyToClipboard(url: string, key: string, quiet = false): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(url);
      copied = key;
      setTimeout(() => {
        if (copied === key) copied = null;
      }, COPY_FLASH_MS);
      return true;
    } catch {
      if (!quiet) error = "clipboard write failed — copy the link from a focused, HTTPS tab";
      return false;
    }
  }

  function linkURL(gameID: string, kind: InviteKind, token: string): string {
    return kind === "player" ? inviteURL(gameID, token) : spectatorInviteURL(gameID, token);
  }

  function heldLinks(kind: InviteKind): SvelteMap<string, string> {
    return kind === "player" ? recentInvites : recentSpectatorInvites;
  }

  // copyLink copies one of the table's links (ADR 0158 §1 item 2,
  // ADR 0212 §4). With none held here, one who may rotate invites gets a
  // new one minted first, as the menu item says. Nobody else is ever
  // offered the item without a link.
  async function copyLink(g: GameMeta, kind: InviteKind): Promise<void> {
    const held = heldLinks(kind);
    let token = held.get(g.id);
    if (!token) {
      if (!canRotateInvites(g)) return;
      rotateBusy = `${g.id}:${kind}`;
      try {
        token = (await rotateInvite(g.id, kind)).token;
        held.set(g.id, token);
      } catch (err) {
        rotateError = {
          gameID: g.id,
          message:
            err instanceof LobbyApiError
              ? err.message
              : `could not make a ${inviteKindLabel(kind)}`,
        };
        return;
      } finally {
        rotateBusy = null;
      }
    }
    await copyToClipboard(linkURL(g.id, kind, token), `${g.id}:${kind}`);
  }

  // The Invite menu's links (ADR 0212 §4): the invite for the host, the
  // creator and the admin in the lobby; the spectator link for anyone at
  // a table not yet over who holds one, or may make one.
  function offersSpectatorLink(g: GameMeta): boolean {
    if (g.state === "ended") return false;
    return recentSpectatorInvites.has(g.id) || canRotateInvites(g);
  }

  function linkTitle(g: GameMeta, kind: InviteKind): string {
    return heldLinks(kind).has(g.id)
      ? `copy this table's ${inviteKindLabel(kind)}`
      : "no link is held here: this makes a new one, and an older link stops working";
  }

  // --- rotating a lost or compromised invite ----------------------
  //
  // Rotating revokes the current invite of one kind and mints a
  // replacement. The old link stops working at once, so it asks first;
  // the new one goes straight to the clipboard (#3499), and only when
  // the browser refuses that does the card show it to copy by hand.
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
      const url = linkURL(gameID, kind, res.token);
      heldLinks(kind).set(gameID, res.token);
      confirmRotate = null;
      if (rotated?.gameID === gameID && rotated.kind === kind) rotated = null;
      if (!(await copyToClipboard(url, `${gameID}:${kind}:new`, true))) {
        rotated = { gameID, kind, url };
      }
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

  // The card's one primary (ADR 0212 §2, ADR 0158 §1 item 2). In the
  // lobby it follows the table's state (tablesPage.lobbyPrimary): Open
  // table, Take a seat, Choose your deck, Waiting for …, Start game or
  // Watch. In play it is Resume for a seat, Open for anyone else; a seat
  // held through another session is reclaimed first.
  interface Primary {
    kind: LobbyPrimaryKind;
    label: string;
    run: () => void;
    disabled?: boolean;
    title?: string;
    busy?: boolean;
  }

  function mayStart(g: GameMeta): boolean {
    return g.state === "lobby" && g.players.length >= 2 && (startsFromSeat(g) || isAdmin);
  }

  // watchesHere: a spectator session of this table (owner decision 6).
  function watchesHere(g: GameMeta): boolean {
    return $session?.principal.role === "spectator" && $session.gameID === g.id;
  }

  function primaryFor(g: GameMeta): Primary {
    const reclaim = (label: string): Primary => ({
      kind: "open",
      label,
      run: () => void openMine(g.id, g.state),
      disabled: opening !== null,
      busy: opening === g.id,
      title: "open your seat at this table",
    });
    if (g.state !== "lobby") {
      if (rejoinsHere(g.id)) return reclaim("Resume");
      const held = mySeatNumber(g) !== null;
      return { kind: "open", label: held ? "Resume" : "Open", run: () => openGame(g.id) };
    }
    const p = lobbyPrimary(g, {
      me: mySeat(g),
      heldElsewhere: rejoinsHere(g.id),
      takesSeat: takesSeat(g),
      mayStart: mayStart(g),
      spectator: watchesHere(g),
      picking: pickerOpen(g) ? (pickerState?.commit ?? null) : null,
      addingBot: botPickerFor === g.id ? botCommit : null,
    });
    switch (p.kind) {
      case "open":
        return rejoinsHere(g.id) ? reclaim(p.label) : { ...p, run: () => openGame(g.id) };
      case "take":
        return {
          ...p,
          label: seatBusy === g.id ? "Taking a seat…" : p.label,
          disabled: seatBusy !== null,
          run: () => void takeSeat(g),
        };
      case "choose":
        return { ...p, run: () => void openPicker(g) };
      case "sit":
        return { ...p, run: () => void deckPicker?.commit() };
      case "bot":
        return { ...p, run: () => void botPicker?.commit() };
      case "start":
        return { ...p, run: () => void onStart(g.id) };
      case "watch":
        return { ...p, run: () => openGame(g.id) };
      default:
        return { ...p, run: () => undefined };
    }
  }

  // Copy invite (§1 item 2): the host or creator, in the lobby, when
  // there is a link to copy or one they may make.
  function offersCopyInvite(g: GameMeta): boolean {
    if (g.state !== "lobby") return false;
    const mayShare = isAdmin || g.is_creator === true || mySeat(g)?.is_host === true;
    return mayShare && (recentInvites.has(g.id) || canRotateInvites(g));
  }

  // takesSeat: the creator, unseated, at their own table in the lobby.
  function takesSeat(g: GameMeta): boolean {
    return g.state === "lobby" && g.is_creator === true && !mySeat(g) && recentInvites.has(g.id);
  }

  function offersDeck(g: GameMeta): boolean {
    return g.state === "lobby" && mySeat(g) !== null && !!$session?.playerID;
  }

  function deckLabel(seat: SeatInfo): string {
    return seat.deck_uploaded
      ? `Change deck · ${seat.deck_name || "deck ready"}`
      : L.chooseYourDeck;
  }

  // tableMenuFor is the card's right-click menu (#3499): every action the
  // card shows, as a shortcut, then the admin's, which only live here:
  // a seat link for each person (ADR 0212 §1), the replay and Archive.
  // Labels are unique within it (they key the list).
  function tableMenuFor(g: GameMeta): MoreItem[] {
    const items: MoreItem[] = [];
    const primary = primaryFor(g);
    if (g.state !== "lobby") {
      items.push({
        label: `${primary.label} table`,
        disabled: primary.disabled,
        title: primary.title,
        onclick: primary.run,
      });
    } else {
      // A wait is not an action; the card says what it waits on.
      if (primary.kind !== "wait") {
        items.push({
          label: primary.label,
          disabled: primary.disabled,
          title: primary.title,
          onclick: primary.run,
        });
      }
      if (primary.kind !== "open" && primary.kind !== "watch") {
        items.push({ label: L.openTable, onclick: () => openGame(g.id) });
      }
    }
    if (takesSeat(g) && primary.kind !== "take") {
      items.push({
        label: seatBusy === g.id ? "Taking a seat…" : "Take a seat",
        disabled: seatBusy !== null,
        onclick: () => void takeSeat(g),
      });
    }
    if (offersCopyInvite(g)) {
      items.push({
        label: L.copyInvite,
        title: linkTitle(g, "player"),
        disabled: rotateBusy !== null,
        onclick: () => void copyLink(g, "player"),
      });
      if (canRotateInvites(g)) {
        items.push({
          label: L.newInviteLink,
          title: "mint a new player invite — the current one stops working immediately",
          disabled: rotateBusy !== null,
          onclick: () => (confirmRotate = { gameID: g.id, kind: "player" }),
        });
      }
    }
    if (offersSpectatorLink(g)) {
      items.push({
        label: L.copySpectatorLink,
        title: linkTitle(g, "spectator"),
        disabled: rotateBusy !== null,
        onclick: () => void copyLink(g, "spectator"),
      });
      if (canRotateInvites(g)) {
        items.push({
          label: L.newSpectatorLink,
          title: "mint a new spectator link — the current one stops working immediately",
          disabled: rotateBusy !== null,
          onclick: () => (confirmRotate = { gameID: g.id, kind: "spectator" }),
        });
      }
    }
    const seat = mySeat(g);
    if (seat && offersDeck(g) && primary.kind !== "choose" && primary.kind !== "sit") {
      items.push({ label: deckLabel(seat), onclick: () => void openPicker(g) });
    }
    if (canInvite(g)) {
      items.push({ label: "Invite by DM…", onclick: () => showPanel(g, "invite") });
    }
    if (canReadSettings(g)) {
      items.push({
        label: settingsFor === g.id ? "Hide table settings" : "Table settings",
        title: "undos, starting life, commander damage, table pace, spawning",
        onclick: () => toggleTableSettings(g.id),
      });
    }
    const lfg = $community.lfg && canManageTableFor(g) ? lfgHeader(g) : null;
    if (lfg?.kind === "post") {
      items.push({
        label: "Post to #looking-for-game",
        title: lfg.title,
        disabled: lfg.disabled || lfgBusy,
        onclick: () => void toggleLFG(g.id, true),
      });
    } else if (lfg?.kind === "posted" && g.lfg && g.state === "lobby") {
      items.push({
        label: "Withdraw from #looking-for-game",
        title:
          "take this table's post down from the CMD & CTRL Discord's #looking-for-game channel",
        disabled: lfgBusy,
        onclick: () => void toggleLFG(g.id, false),
      });
    }
    // Admin-only: a disconnected player's way back in, one per person.
    if (isAdmin) {
      for (const p of [...g.players].sort((a, b) => a.seat - b.seat)) {
        if (p.is_bot) continue;
        items.push({
          label: `Seat link for ${seatName(p)} (seat ${p.seat + 1})`,
          title: `generate a one-time link that returns ${seatName(p)} to this seat`,
          disabled: reclaimBusy !== null,
          onclick: () => void onSeatLink(g.id, p),
        });
      }
    }
    // The server's replay gate: an admin any time after the lobby.
    if (isAdmin && g.state !== "lobby") {
      const url = replayURL(g.id);
      if (url) items.push({ label: "Download replay", href: url, download: `${g.id}.jsonl` });
    }
    if (isAdmin) {
      items.push({
        label: "Archive…",
        title: "hide this table from the list — nothing is deleted",
        disabled: manageBusy,
        onclick: () => (confirming = { id: g.id, kind: "archive" }),
      });
    }
    return items;
  }

  // The open card menu: which table, where, and what to give focus back
  // to when Escape closes it.
  let cardMenu = $state<{ gameID: string; at: MenuPoint; opener: HTMLElement | null } | null>(null);
  const cardMenuGame = $derived(
    cardMenu ? (games.find((g) => g.id === cardMenu!.gameID) ?? null) : null,
  );

  function openCardMenu(g: GameMeta, at: MenuPoint, opener: HTMLElement | null): void {
    cardMenu = { gameID: g.id, at, opener };
  }

  function closeCardMenu(restoreFocus: boolean): void {
    const opener = cardMenu?.opener;
    cardMenu = null;
    if (restoreFocus) opener?.focus();
  }

  // --- the Invite menu (ADR 0212 §4) --------------------------------
  //
  // One menu for the footer's Invite ▾ and each open seat's Invite: the
  // two links, Invite by DM, #looking-for-game, and under "Make new
  // links" the two rotations with today's confirm. It shows only what
  // this person may use, and every item is in the right-click menu too.
  // `from` is the button that opened it: "footer", or an open seat's
  // number.
  let inviteMenu = $state<{
    gameID: string;
    from: "footer" | number;
    opener: HTMLElement;
    x: number;
    y: number;
    align: "start" | "end";
  } | null>(null);
  const inviteMenuGame = $derived(
    inviteMenu ? (games.find((g) => g.id === inviteMenu!.gameID) ?? null) : null,
  );
  // The press that closes the menu from outside lands on its own button
  // first; this keeps that button's click from opening it again.
  let inviteClosed: { opener: HTMLElement | undefined; at: number } | null = null;

  function toggleInviteMenu(g: GameMeta, opener: HTMLElement, from: "footer" | number): void {
    if (inviteMenu?.opener === opener) {
      closeInviteMenu(false);
      return;
    }
    if (inviteClosed?.opener === opener && performance.now() - inviteClosed.at < 400) return;
    const r = opener.getBoundingClientRect();
    const align = from === "footer" ? "end" : "start";
    inviteMenu = {
      gameID: g.id,
      from,
      opener,
      x: align === "end" ? r.right : r.left,
      y: r.bottom + 6,
      align,
    };
  }

  function closeInviteMenu(restoreFocus: boolean): void {
    const opener = inviteMenu?.opener;
    inviteClosed = { opener, at: performance.now() };
    inviteMenu = null;
    if (restoreFocus) opener?.focus();
  }

  function inviteItemsFor(g: GameMeta): MoreItem[] {
    // A choice gives focus back to the button that opened the menu.
    const back =
      (run: () => void): (() => void) =>
      () => {
        const opener = inviteClosed?.opener;
        run();
        opener?.focus();
      };
    const items: MoreItem[] = [];
    const unheld = "No link is held here: this makes a new one, and an older one stops working";
    if (offersCopyInvite(g)) {
      items.push({
        label: L.copyInvite,
        icon: "link",
        hint: recentInvites.has(g.id) ? "The player link: whoever opens it takes a seat" : unheld,
        disabled: rotateBusy !== null,
        onclick: back(() => void copyLink(g, "player")),
      });
    }
    if (offersSpectatorLink(g)) {
      items.push({
        label: L.copySpectatorLink,
        icon: "eye",
        hint: recentSpectatorInvites.has(g.id) ? "Watch only; no seat" : unheld,
        disabled: rotateBusy !== null,
        onclick: back(() => void copyLink(g, "spectator")),
      });
    }
    if (canInvite(g)) {
      items.push({
        label: "Invite by DM…",
        icon: "mail",
        hint: "Someone you've played with, on Discord",
        onclick: back(() => showPanel(g, "invite")),
      });
    }
    const lfg = $community.lfg && canManageTableFor(g) ? lfgHeader(g) : null;
    if (lfg?.kind === "post") {
      items.push({
        label: "Post to #looking-for-game",
        icon: "discord",
        hint: lfg.disabled ? lfg.title : "Comes down when the game starts",
        disabled: lfg.disabled || lfgBusy,
        onclick: back(() => void toggleLFG(g.id, true)),
      });
    } else if (lfg?.kind === "posted") {
      items.push({
        label: "Withdraw from #looking-for-game",
        icon: "discord",
        hint: `${lfg.label}; this takes the post down`,
        disabled: lfgBusy,
        onclick: back(() => void toggleLFG(g.id, false)),
      });
    }
    if (canRotateInvites(g)) {
      const group = "Make new links";
      if (offersCopyInvite(g)) {
        items.push({
          label: L.newInviteLink,
          icon: "newLink",
          hint: "The current player link stops working",
          group,
          disabled: rotateBusy !== null,
          onclick: back(() => (confirmRotate = { gameID: g.id, kind: "player" })),
        });
      }
      if (offersSpectatorLink(g)) {
        items.push({
          label: L.newSpectatorLink,
          icon: "newLink",
          hint: "The current spectator link stops working",
          group,
          disabled: rotateBusy !== null,
          onclick: back(() => (confirmRotate = { gameID: g.id, kind: "spectator" })),
        });
      }
    }
    return items;
  }

  // invitesPlayers: the Invite menu can fill a seat for this person (the
  // player link, a DM or #lfg), so an open seat offers it too. Someone
  // who may only share the spectator link finds it in the footer.
  function invitesPlayers(g: GameMeta): boolean {
    const lfg = $community.lfg && canManageTableFor(g) ? lfgHeader(g) : null;
    return offersCopyInvite(g) || canInvite(g) || lfg !== null;
  }

  // What the Invite button says for a moment after a copy, where a
  // screen reader hears it once (#3499's split buttons said it in place).
  function copiedNote(gameID: string): string {
    switch (copied) {
      case `${gameID}:player`:
        return "Invite copied";
      case `${gameID}:player:new`:
        return "New invite copied";
      case `${gameID}:spectator`:
        return "Spectator link copied";
      case `${gameID}:spectator:new`:
        return "New spectator link copied";
      default:
        return "";
    }
  }

  // The panels under the seats: invite a tablemate (from the Invite
  // menu), and your deck (your seat's tile, or Choose your deck). The
  // deck picker opens on its own while your seat has no deck, the table
  // you just made included (ADR 0212 owner decision 2); Invite by DM
  // opens first on a new table where you are not seated. Either stays
  // as the person last left it. They share the region under the seats
  // with the bot picker, so opening one closes it.
  let panelChoice = $state<Record<string, "invite" | "deck" | null>>({});

  function panelFor(g: GameMeta): "invite" | "deck" | null {
    if (g.id in panelChoice) return panelChoice[g.id];
    const seat = mySeat(g);
    if (seat && g.state === "lobby" && !seat.deck_uploaded) return "deck";
    if (justCreated === g.id) return "invite";
    return null;
  }

  function togglePanel(g: GameMeta, which: "invite" | "deck"): void {
    const next = panelFor(g) === which ? null : which;
    if (next && botPickerFor === g.id) botPickerFor = null;
    panelChoice = { ...panelChoice, [g.id]: next };
  }

  // showPanel opens one from a menu, where a second click closing it
  // again would be a surprise.
  function showPanel(g: GameMeta, which: "invite" | "deck"): void {
    if (botPickerFor === g.id) botPickerFor = null;
    panelChoice = { ...panelChoice, [g.id]: which };
  }

  // --- the deck picker (ADR 0212 §3) --------------------------------
  //
  // One at most is open, since a session holds one seat. It reports what
  // is selected, which your seat's tile previews, and what committing it
  // says, which the footer's primary becomes (§2 rule 4): "Sit down with
  // this deck", or Upload / Import deck on Paste a list. The primary
  // calls its commit().
  let pickerState = $state<PickerState | null>(null);
  let deckPicker = $state<{ commit: () => Promise<void> } | undefined>();

  function pickerOpen(g: GameMeta): boolean {
    return panelFor(g) === "deck" && offersDeck(g);
  }

  // After a deck is seated the picker closes and your tile says Ready,
  // unless the picker has more to say about the deck just seated (its
  // warnings, or a pasted list's unimplemented cards and library note),
  // which stays open to be read. The answer carries the table's seats,
  // so the tile turns Ready before the list is read again.
  function onSeated(g: GameMeta, res: UploadDeckResponse, more: boolean): void {
    const seats = res.game?.players;
    if (seats) games = games.map((x) => (x.id === g.id ? { ...x, players: seats } : x));
    panelChoice = { ...panelChoice, [g.id]: more ? "deck" : null };
    void refresh();
  }

  // The deck picker's id, for the seat tile's aria-controls.
  function pickerID(gameID: string): string {
    return `deck-picker-${gameID}`;
  }

  // openPicker is Choose your deck: open the deck picker, or bring the
  // open one into view, and put focus on it.
  async function openPicker(g: GameMeta): Promise<void> {
    if (panelFor(g) !== "deck") showPanel(g, "deck");
    await tick();
    const panel = document.getElementById(pickerID(g.id));
    panel?.scrollIntoView?.({ block: "nearest" });
    panel?.focus({ preventScroll: true });
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

  // A table in play shows who is there, in seat order; a table in the
  // lobby draws its four seats as tiles (tablesPage.lobbySlots).
  function seatsInPlay(g: GameMeta): SeatInfo[] {
    return [...g.players].sort((a, b) => a.seat - b.seat);
  }

  // A bot seat's pre-built deck, from GET /bot/options.
  function botDeckOf(p: SeatInfo): BotDeckInfo | undefined {
    if (!p.is_bot || !p.bot_deck) return undefined;
    return botOptions?.decks?.find((d) => d.id === p.bot_deck);
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

  // "Cards from this game" (ADR 0179 §7): the ended games whose panel is
  // open. Offered only for an ended game the person sat at, and only
  // where they see shop links; the server refuses a game still running.
  let cardsOpen = $state<string[]>([]);
  function toggleCards(id: string): void {
    cardsOpen = cardsOpen.includes(id) ? cardsOpen.filter((x) => x !== id) : [...cardsOpen, id];
  }

  function historyMore(r: HistoryRow): MoreItem[] {
    const items: MoreItem[] = [];
    if (r.state === "ended" && r.seated && $visibleShops.length > 0) {
      items.push({
        label: cardsOpen.includes(r.id) ? "Hide cards from this game" : "Cards from this game",
        onclick: () => toggleCards(r.id),
      });
    }
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
    void fetchPrebuiltDecks()
      .then((r) => (prebuiltDecks = r.decks ?? []))
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
        {#if $community.lfg}
          <label class="use-setup">
            <input type="checkbox" bind:checked={wantLFG} disabled={lfgCreateBlocked} />
            <span>
              Post to the CMD & CTRL Discord (#looking-for-game)
              <span class="setup-sum"
                >{lfgCreateBlocked
                  ? LFG_GUEST_HINT
                  : "the table's name, you as host and the open seats, so players can join"}</span
              >
            </span>
          </label>
        {/if}
        {#if $community.results}
          <label class="use-setup">
            <input type="checkbox" bind:checked={wantResults} />
            <span>
              Post the result to the CMD & CTRL Discord
              <span class="setup-sum"
                >when the game ends: the table's name, seat names, commanders and who won</span
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
            {@const lobby = g.state === "lobby"}
            {@const seat = mySeat(g)}
            {@const myNo = mySeatNumber(g)}
            {@const chip = statusChip(g)}
            {@const primary = primaryFor(g)}
            {@const panel = panelFor(g)}
            {@const rules = cardSettings.get(g.id)}
            {@const rulesText = rules ? settingsSummary(rules) : null}
            {@const showRules = canReadSettings(g)}
            {@const inviteItems = inviteItemsFor(g)}
            {@const footerMenuOpen = inviteMenu?.gameID === g.id && inviteMenu.from === "footer"}
            <li
              class="tcard"
              class:lobby
              class:mine={myNo !== null}
              class:menu-open={cardMenu?.gameID === g.id}
              use:contextMenuTrigger={(at, opener) => openCardMenu(g, at, opener)}
            >
              <div class="thead">
                <div class="tid">
                  <h3 class="tname">{g.name}</h3>
                  <span class="chip" class:live={chip.live}>
                    {#if chip.live}<i class="dot ok"></i>{/if}{chip.label}
                  </span>
                  <span class="ctx">{contextLine(g, { seat: myNo })}</span>
                </div>
                {#if lobby}
                  <!-- §5: the settings as one quiet pill. -->
                  {#if showRules}
                    <button
                      type="button"
                      class="rules pill"
                      aria-label={L.tableSettingsLine(rulesText ?? "loading")}
                      aria-expanded={settingsFor === g.id}
                      title={canManageTableFor(g)
                        ? "table settings: undos, starting life, commander damage, table pace, spawning"
                        : "the table's settings; only the host can change them"}
                      onclick={() => toggleTableSettings(g.id)}
                    >
                      <Icon name="gear" size={13} />
                      <span>{rulesText ?? "Table settings"}</span>
                      <i class="caret" class:up={settingsFor === g.id}></i>
                    </button>
                  {/if}
                {:else}
                  <div class="tact">
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
                {/if}
              </div>

              {#if lobby}
                <!-- §1: the seats are the hero, four tiles at every width. -->
                <ul class="tiles" aria-label="seats">
                  {#each lobbySlots(g) as p, i (p?.player_id ?? `open-${i}`)}
                    {#if p}
                      {@const own = !p.is_bot && isOwnSeat(p, { isAdmin, mySeat: seat })}
                      <!-- Your seat previews the deck selected in the open picker (§1). -->
                      {@const pv =
                        seat?.player_id === p.player_id && pickerOpen(g)
                          ? seatPreview(p, pickerState?.preview ?? null)
                          : null}
                      <SeatTile
                        seat={p.seat}
                        {p}
                        text={pv?.text ?? seatTileText(p, botDeckOf(p))}
                        art={pv?.art ?? []}
                        preview={pv?.preview ?? false}
                        you={myNo === p.seat}
                        botChip={p.is_bot ? botChipText(p, botOptions?.tiers ?? []) : ""}
                        pick={seat?.player_id === p.player_id && offersDeck(g)
                          ? {
                              expanded: panel === "deck",
                              controls: pickerID(g.id),
                              onclick: () => togglePanel(g, "deck"),
                            }
                          : undefined}
                        remove={(p.is_bot ? canManageBots(g) : canRemove(g, p))
                          ? {
                              label: own ? "leave your seat" : `remove ${seatName(p)}`,
                              title: own
                                ? "leave your seat at this table"
                                : `remove ${seatName(p)} from this table`,
                              disabled: removeBusy || botBusy,
                              onclick: () => askRemove(g, p),
                            }
                          : undefined}
                      />
                    {:else}
                      <SeatTile
                        seat={i}
                        p={null}
                        oninvite={invitesPlayers(g)
                          ? (el) => toggleInviteMenu(g, el, i)
                          : undefined}
                        inviteExpanded={inviteMenu?.gameID === g.id && inviteMenu.from === i}
                        addBot={canManageBots(g) && botPickerFor !== g.id
                          ? { disabled: botBusy, onclick: () => void openBotPicker(g) }
                          : undefined}
                      />
                    {/if}
                  {/each}
                </ul>
              {:else}
                <ul class="seats" aria-label="seats">
                  {#each seatsInPlay(g) as p (p.player_id)}
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
                            ></i>{/if}<DeckPowerBadge power={line.power} /><span class="stext"
                            >{line.text}</span
                          >
                        </div>
                      </div>
                      {#if canRemove(g, p)}
                        <button
                          class="seat-x"
                          title={`remove ${seatName(p)} from this table`}
                          aria-label={`remove ${seatName(p)}`}
                          disabled={removeBusy}
                          onclick={() => askRemove(g, p)}
                        >
                          <Icon name="x" size={12} />
                        </button>
                      {/if}
                    </li>
                  {/each}
                </ul>
              {/if}

              {#if lfgError && lfgErrorFor === g.id && settingsFor !== g.id}
                <div class="bot-error" role="alert">{lfgError}</div>
              {/if}

              {#if !lobby && (showRules || inviteItems.length > 0)}
                <!-- In play: the rules line, and the Invite menu for the
                     spectator link. -->
                <div class="tline">
                  {#if showRules}
                    <button
                      type="button"
                      class="rules"
                      aria-label={L.tableSettingsLine(rulesText ?? "loading")}
                      aria-expanded={settingsFor === g.id}
                      title={canManageTableFor(g)
                        ? "table settings: undos, starting life, commander damage, table pace, spawning"
                        : "the table's settings; only the host can change them"}
                      onclick={() => toggleTableSettings(g.id)}
                    >
                      <Icon name="gear" size={13} />
                      <span>{rulesText ?? "Table settings"}</span>
                      <i class="caret" class:up={settingsFor === g.id}></i>
                    </button>
                  {/if}
                  {#if inviteItems.length > 0}
                    <span class="tacts">
                      <span class="copied" role="status">{copiedNote(g.id)}</span>
                      <button
                        type="button"
                        class="btn ghost sm"
                        aria-haspopup="menu"
                        aria-expanded={footerMenuOpen}
                        onclick={(e) => toggleInviteMenu(g, e.currentTarget, "footer")}
                      >
                        <Icon name="link" size={13} />
                        {L.inviteButton}
                        <i class="caret" class:up={footerMenuOpen}></i>
                      </button>
                    </span>
                  {/if}
                </div>
              {/if}

              <!-- The region under the seats: the bot picker (owner decision
                   8), the tablemate picker, or the deck picker (§3), with a
                   notch at the seat it is for. -->
              {#if botPickerFor === g.id && lobby && botOptions}
                {@const botSeat = botSeatFor(g)}
                <div
                  class="picker bot"
                  id={botPickerID(g.id)}
                  tabindex="-1"
                  style:--notch={Math.min(botSeat, 3)}
                >
                  <BotPicker
                    bind:this={botPicker}
                    gameID={g.id}
                    seat={botSeat}
                    options={botOptions}
                    prebuilt={prebuiltDecks}
                    onchange={(c) => (botCommit = c)}
                    oncancel={() => (botPickerFor = null)}
                    onadded={() => {
                      botPickerFor = null;
                      void refresh();
                    }}
                  />
                </div>
              {:else if panel === "invite" && canInvite(g)}
                <div class="invite-picker">
                  <TablematePicker gameID={g.id} lastTable={lastSetup?.tablemates ?? []} />
                </div>
              {:else if panel === "deck" && seat && offersDeck(g) && $session?.playerID}
                <div
                  class="picker"
                  id={pickerID(g.id)}
                  role="region"
                  aria-label={L.deckPicker}
                  tabindex="-1"
                  style:--notch={seat.seat}
                >
                  <DeckPicker
                    bind:this={deckPicker}
                    gameID={g.id}
                    playerID={$session.playerID}
                    shopLinks
                    actions="external"
                    onchange={(st) => (pickerState = st)}
                    onSuccess={(res, more) => onSeated(g, res, more)}
                  />
                </div>
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
                  {#if $community.lfg}
                    <LfgPost
                      game={g}
                      canManage={canManageTableFor(g)}
                      ontoggle={(enabled) => toggleLFG(g.id, enabled)}
                      busy={lfgBusy}
                      error={lfgError}
                    />
                  {/if}
                  {#if $community.results}
                    <ResultsPost
                      game={g}
                      canManage={canManageTableFor(g)}
                      ontoggle={(enabled) => toggleResults(g.id, enabled)}
                      busy={resultsBusy}
                      error={resultsError}
                    />
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
                {@const rk = confirmRemove.kind}
                <div class="confirm danger" role="alert">
                  <p>
                    <strong
                      >{rk === "self"
                        ? LEAVE_SEAT_QUESTION
                        : removeSeatQuestion(seatName(confirmRemove.seat))}</strong
                    >
                    {rk === "self"
                      ? leaveSeatConsequence(g, confirmRemove.seat.is_host === true)
                      : rk === "bot"
                        ? REMOVE_BOT_CONSEQUENCE
                        : removeSeatConsequence(g)}
                  </p>
                  {#if rk === "person"}
                    <label class="block-opt">
                      <input type="checkbox" bind:checked={removeBlock} disabled={removeBusy} />
                      <span>{BLOCK_LABEL}</span>
                    </label>
                    <p class="hint">{BLOCK_HINT}</p>
                  {/if}
                  <div class="confirm-actions">
                    <button
                      class="primary"
                      disabled={removeBusy}
                      onclick={() => onRemoveSeat(g.id, confirmRemove!.seat.player_id, rk)}
                    >
                      {removeBusy
                        ? "removing…"
                        : rk === "self"
                          ? "leave my seat"
                          : "remove from table"}
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

              {#if lobby}
                <!-- §2: the readiness line, then Open table, the Invite
                     menu and the one primary, in the footer's corner. -->
                {@const ready = readiness(g, seat, { bots: canManageBots(g) })}
                <div class="tfoot">
                  <div
                    class="readiness"
                    role="group"
                    aria-label={L.readinessLine(ready.ready, ready.seated)}
                  >
                    <span class="rlabel" aria-hidden="true">Ready</span>
                    <span class="rcount" aria-hidden="true">{ready.ready}/{ready.seated}</span>
                    <span class="segs" aria-hidden="true">
                      {#each ready.segments as sg (sg.seat)}
                        <i class="sg {sg.state}" style:--c={seatColor(sg.seat)}></i>
                      {/each}
                    </span>
                    <span class="rtext"
                      >{ready.lead ? `${ready.lead} · ` : ""}<b>{ready.status}</b></span
                    >
                  </div>
                  <div class="tacts">
                    <span class="copied" role="status">{copiedNote(g.id)}</span>
                    {#if primary.kind !== "open" && primary.kind !== "watch"}
                      <button type="button" class="btn ghost sm" onclick={() => openGame(g.id)}>
                        {L.openTable}
                      </button>
                    {/if}
                    {#if inviteItems.length > 0}
                      <button
                        type="button"
                        class="btn sm"
                        aria-haspopup="menu"
                        aria-expanded={footerMenuOpen}
                        onclick={(e) => toggleInviteMenu(g, e.currentTarget, "footer")}
                      >
                        <Icon name="link" size={13} />
                        {L.inviteButton}
                        <i class="caret" class:up={footerMenuOpen}></i>
                      </button>
                    {/if}
                    <button
                      type="button"
                      class="btn primary sm wide"
                      disabled={primary.disabled}
                      title={primary.title}
                      onclick={primary.run}
                    >
                      {primary.busy ? "…" : primary.label}
                    </button>
                  </div>
                </div>
              {/if}
            </li>
          {/each}
        </ul>
        {#if cardMenu && cardMenuGame}
          {#key cardMenu}
            <ContextMenu
              label={L.tableMenu(cardMenuGame.name)}
              x={cardMenu.at.x}
              y={cardMenu.at.y}
              items={tableMenuFor(cardMenuGame)}
              onclose={closeCardMenu}
            />
          {/key}
        {/if}
        {#if inviteMenu && inviteMenuGame}
          {#key inviteMenu}
            <ContextMenu
              label={L.inviteMenu(inviteMenuGame.name)}
              x={inviteMenu.x}
              y={inviteMenu.y}
              align={inviteMenu.align}
              items={inviteItemsFor(inviteMenuGame)}
              onclose={closeInviteMenu}
            />
          {/key}
        {/if}
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
                {#if r.state === "ended" && cardsOpen.includes(r.id)}
                  <div class="hconfirm">
                    <GameCardsPanel gameID={r.id} />
                  </div>
                {/if}
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

  /* Four seats in one row at every width, the tablet's 820px included
     (ADR 0212 §6): the lobby's tiles and a game in play's seats. */
  .seats,
  .tiles {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px;
  }
  .tiles {
    gap: 12px;
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
  /* A seat in play: four across at 820px too (ADR 0212 §6), so the
     line wraps rather than cut off its life total. The dot and the power
     badge beside it never shrink (#3671). */
  .sstat {
    font-size: 11.5px;
    color: var(--fg-muted);
    margin-top: 2px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 2px 5px;
  }
  .sstat.pend {
    color: var(--accent-strong);
  }
  /* Beside the dot when there is room for a word or two, else under it
     and the badge. */
  .stext {
    flex: 1 1 4.5rem;
    min-width: 0;
    overflow-wrap: break-word;
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
  .seat-x {
    display: grid;
    place-items: center;
    flex: 0 0 auto;
    width: 22px;
    height: 22px;
    padding: 0;
    border-radius: 8px;
    border: 1px solid var(--border-strong);
    background: transparent;
    color: var(--fg-muted);
    cursor: pointer;
  }
  .seat-x:hover:not(:disabled) {
    color: var(--fg);
    border-color: var(--accent);
  }
  .seat-x:disabled {
    opacity: 0.5;
    cursor: default;
  }

  /* --- the settings summary, and in play the rules line (#3499) --- */
  .tcard.menu-open {
    border-color: var(--border-strong);
  }
  .tline {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 16px;
    min-width: 0;
  }
  .tline .tacts {
    margin-left: auto;
  }
  .rules {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    max-width: 100%;
    padding: 2px 0;
    border: 0;
    background: transparent;
    color: var(--fg-muted);
    font: 600 12.5px/1.4 var(--font-ui);
    text-align: left;
    cursor: pointer;
  }
  .rules span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .rules:hover {
    color: var(--fg);
  }
  /* §5: in the lobby card's head, one quiet pill. */
  .rules.pill {
    flex: 0 1 auto;
    height: 28px;
    padding: 0 11px;
    border: 1px solid var(--border);
    border-radius: 999px;
    font-weight: 500;
  }
  .rules.pill:hover {
    border-color: var(--border-strong);
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
  .invite-picker {
    padding: 12px 14px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface-sunken);
  }
  /* The picker region under the seats (§3): the deck picker, its tab bar
     edge to edge, or the bot picker. Its notch points at the seat it is
     for: the middle of column --notch of four, 12px apart. */
  .picker {
    --picker-pad: 14px;
    position: relative;
    padding: 0 var(--picker-pad) var(--picker-pad);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    background: var(--surface-sunken);
  }
  .picker.bot {
    padding-top: 12px;
  }
  .picker::before {
    content: "";
    position: absolute;
    top: -7px;
    left: calc((100% - 36px) / 4 * (var(--notch, 0) + 0.5) + 12px * var(--notch, 0) - 6px);
    width: 12px;
    height: 12px;
    transform: rotate(45deg);
    background: var(--surface-sunken);
    border-left: 1px solid var(--border-strong);
    border-top: 1px solid var(--border-strong);
  }
  .picker:focus {
    outline: none;
  }
  .picker:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }

  /* --- the lobby card's footer: readiness and the one primary (§2) --- */
  .tfoot {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 10px 16px;
    padding-top: 14px;
    border-top: 1px solid var(--border);
  }
  .readiness {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 12px;
    flex: 1 1 260px;
    min-width: 0;
  }
  .rlabel {
    font: 700 10.5px/1 var(--font-mono);
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .rcount {
    margin-left: -6px;
    font: 700 15px/1 var(--font-mono);
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .segs {
    display: inline-flex;
    gap: 4px;
  }
  .sg {
    display: block;
    width: 20px;
    height: 6px;
    box-sizing: border-box;
    border-radius: 3px;
  }
  .sg.ready {
    background: var(--c);
  }
  .sg.choosing {
    border: 1.5px solid var(--c);
  }
  .sg.open {
    border: 1px dashed var(--fg-dim);
  }
  .rtext {
    min-width: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .rtext b {
    color: var(--fg);
    font-weight: 600;
  }
  .tacts {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
  }
  .tacts .btn {
    gap: 6px;
  }
  .btn.wide {
    padding: 0 18px;
  }
  .copied {
    font-size: 12px;
    color: var(--mint);
    white-space: nowrap;
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
