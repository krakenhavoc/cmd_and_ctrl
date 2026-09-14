<script lang="ts">
  import { onMount } from "svelte";
  import {
    addBotSeat,
    archiveGame,
    createGame,
    deleteGame,
    fetchBotOptions,
    listGames,
    logout as apiLogout,
    mintSeatReclaim,
    removeBotSeat,
    replayURL,
    startGame,
    unarchiveGame,
    type BotOptions,
    type GameMeta,
    type ReclaimTicket,
    type SeatInfo,
  } from "../lib/api";
  import { inviteURL, reclaimURL, spectatorInviteURL, navigate } from "../lib/router";
  import { session, LobbyApiError } from "../lib/session";
  import { openSettings } from "../lib/settings";
  import { seatColor } from "../lib/colors";
  import { avatarURL } from "../lib/api";
  import DeckUploadForm from "../lib/components/DeckUploadForm.svelte";
  import Icon from "../lib/components/Icon.svelte";

  // Lobby is the admin + player landing page. Admins see a create-
  // game form and the invite token for each game they've created;
  // players see the game they're seated in with a deck-upload panel
  // for their own seat and a button to jump into the game view once
  // all seats are ready.
  let games = $state<GameMeta[]>([]);
  let error = $state("");
  let newName = $state("");
  let busy = $state(false);

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
    if ($session?.principal.role === "admin") return true;
    return mySeat(g) !== null;
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
  const isAdmin = $derived($session?.principal.role === "admin");
  let showArchived = $state(false);
  let archived = $state<GameMeta[]>([]);
  let manageBusy = $state(false);
  // The pending confirmation, if any: which table and which action.
  let confirming = $state<{ id: string; kind: "archive" | "delete" } | null>(null);

  async function refresh(): Promise<void> {
    try {
      games = await listGames();
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
      const meta = await createGame(newName.trim());
      if (meta.invite_token) recentInvites.set(meta.id, meta.invite_token);
      if (meta.spectator_invite) recentSpectatorInvites.set(meta.id, meta.spectator_invite);
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
      error = "no invite token cached for this game — re-open as admin to recover";
      return;
    }
    await copyToClipboard(inviteURL(id, token), `${id}:invite`);
  }

  async function copySpectatorInvite(id: string): Promise<void> {
    const token = recentSpectatorInvites.get(id);
    if (!token) {
      error = "no spectator invite cached for this game — re-open as admin to recover";
      return;
    }
    await copyToClipboard(spectatorInviteURL(id, token), `${id}:spectator`);
  }

  function openGame(id: string): void {
    navigate(`#/games/${id}`);
  }

  async function logout(): Promise<void> {
    await apiLogout();
    navigate("#/login");
  }

  // mySeat returns the seat this user occupies in game g, or null if
  // the session isn't bound to a seat in g (admin viewing someone
  // else's game, or a RolePlayer viewing a different game entirely).
  function mySeat(g: GameMeta) {
    const s = $session;
    if (!s?.playerID || s.gameID !== g.id) return null;
    return g.players.find((p) => p.player_id === s.playerID) ?? null;
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
    return me ? "Waiting for the admin to start" : "";
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

  // Players see their own table only (the list endpoint returns
  // every game); admins and spectators see the whole room. Falls
  // back to the full list if the seated game isn't in it.
  const visibleGames = $derived.by(() => {
    const s = $session;
    if (s?.principal.role !== "player" || !s.gameID) return games;
    const mine = games.filter((g) => g.id === s.gameID);
    return mine.length > 0 ? mine : games;
  });

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
</script>

<section class="lobby">
  <header class="bar">
    <span class="wordmark" aria-hidden="true"><i></i>CMD &amp; CTRL</span>
    <h1 class="crumb" aria-label="cmd_and_ctrl · lobby"><b>/</b> Tables</h1>
    <span class="bar-spacer"></span>
    <span class="uchip">
      {#if $session?.principal.name && $session.principal.name !== $session.principal.role}
        {$session.principal.name}
      {/if}
      <b>{$session?.principal.role}</b>
    </span>
    <button
      class="ibtn"
      title="settings (press , from anywhere)"
      aria-label="open settings"
      onclick={openSettings}><Icon name="gear" size={17} /></button
    >
    <button class="ghost" onclick={logout}>log out</button>
  </header>

  <div class="head">
    <div>
      <h2 class="title">Tables</h2>
      <p class="sub">{summary}</p>
    </div>
    {#if $session?.principal.role === "admin"}
      <form class="create" onsubmit={onCreate}>
        <h2 class="panel-h">create game</h2>
        <input type="text" placeholder="game name" bind:value={newName} />
        <button type="submit" class="primary" disabled={busy || !newName.trim()}>create</button>
      </form>
    {/if}
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if visibleGames.length === 0}
    <p class="muted empty">
      {#if $session?.principal.role === "admin"}
        Create a table, then send the invite link to your pod.
      {:else}
        You're not seated at a table yet — ask for an invite link.
      {/if}
    </p>
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
                <!-- Mirror the server's downloadReplay gate: admins may
                     pull the replay any time after the lobby phase, but
                     players get 403 until the game has ended (the JSONL
                     carries unfiltered hidden information mid-game). -->
                {#if g.state === "ended" || ($session?.principal.role === "admin" && g.state !== "lobby")}
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
                {#if g.state === "lobby" && seat}
                  <!-- Seated players wait for the admin; the button goes
                       live once the poll sees the table start. -->
                  <button class="primary" disabled title="waiting for the admin to start">
                    enter table <Icon name="chevronRight" size={13} />
                  </button>
                {:else if g.state === "lobby"}
                  <button onclick={() => openGame(g.id)}>open table</button>
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
  /* Same command bar as the game route, minus the game controls. */
  .bar {
    display: flex;
    align-items: center;
    gap: 12px;
    height: 44px;
    margin: -1.5rem -1.5rem 8px;
    padding: 0 14px;
    background: var(--bg-1);
    border-bottom: 1px solid var(--border);
  }
  .wordmark {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 14px;
    letter-spacing: 0.18em;
    color: var(--fg);
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .wordmark i {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid var(--gold);
    transform: rotate(45deg);
    border-radius: 3px;
    box-sizing: border-box;
  }
  h1.crumb {
    margin: 0;
    font-family: var(--font-ui);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0;
    color: var(--fg-muted);
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .crumb b {
    color: var(--fg-dim);
    font-weight: 400;
  }
  .bar-spacer {
    flex: 1;
  }
  .uchip {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 30px;
    padding: 0 10px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--surface);
    font-size: 12.5px;
    font-weight: 600;
    color: var(--fg);
  }
  .uchip b {
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--gold-strong);
    font-weight: 700;
  }
  .ibtn {
    width: 32px;
    height: 32px;
    padding: 0;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--fg-muted);
  }
  .ibtn:hover {
    color: var(--fg);
    background: rgba(255, 255, 255, 0.06);
    border-color: var(--border);
  }

  .head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 20px;
    flex-wrap: wrap;
  }
  h2.title {
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
  .create {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .create input {
    width: 240px;
    margin: 0;
    height: 36px;
    padding: 0 12px;
    box-sizing: border-box;
    font-size: 13px;
  }
  .create button {
    height: 36px;
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

  .deck-upload {
    border-top: 1px solid var(--border);
    padding-top: 12px;
  }
  .deck-upload summary {
    cursor: pointer;
    list-style: none;
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .deck-upload summary::-webkit-details-marker {
    display: none;
  }
  .deck-upload summary::before {
    content: "";
    width: 6px;
    height: 6px;
    border-right: 1.5px solid var(--fg-dim);
    border-bottom: 1.5px solid var(--fg-dim);
    transform: rotate(-45deg);
    transition: transform 120ms var(--ease);
  }
  .deck-upload[open] summary::before {
    transform: rotate(45deg);
  }
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
