<script lang="ts">
  import { onMount } from "svelte";
  import {
    createGame,
    listGames,
    logout as apiLogout,
    startGame,
    uploadDeck,
    type GameMeta,
  } from "../lib/api";
  import { inviteURL, navigate } from "../lib/router";
  import { session, LobbyApiError, type ApiViolation } from "../lib/session";

  // Lobby is the admin + player landing page. Admins see a create-
  // game form and the invite token for each game they've created;
  // players see the game they're seated in with a deck-upload panel
  // for their own seat and a button to jump into the game view once
  // all seats are ready.
  let games = $state<GameMeta[]>([]);
  let error = $state("");
  let newName = $state("");
  let busy = $state(false);

  // Track the freshly-created game's invite token client-side — the
  // List endpoint strips invite tokens, so we remember them per
  // session so admins can copy the link without re-fetching /games/{id}.
  const recentInvites = new Map<string, string>();

  // Per-game deck-upload form state. Keyed by game ID so feedback
  // for one upload doesn't bleed into the panel for another game.
  const deckSources = $state<Record<string, string>>({});
  const deckErrors = $state<Record<string, string>>({});
  const deckViolations = $state<Record<string, ApiViolation[]>>({});
  const deckWarnings = $state<Record<string, ApiViolation[]>>({});
  const deckSuccess = $state<Record<string, string>>({});
  let deckBusy = $state("");

  // violationLabel prefixes the card name (when present) onto the
  // server message so the UI renders a compact one-line-per-failure
  // list without needing a table.
  function violationLabel(v: ApiViolation): string {
    return v.card ? `${v.card}: ${v.message}` : v.message;
  }

  // looksLikeURL flips the upload-button label + spinner text when
  // the pasted source is clearly a URL, so the user sees
  // "fetching from moxfield.com…" instead of a generic "uploading…"
  // while an outbound request is in flight. Matches the server's
  // auto-detect heuristic (leading http:// or https://) so the
  // button text doesn't contradict what the server will actually do.
  function looksLikeURL(source: string): boolean {
    const s = source.trim();
    return s.startsWith("http://") || s.startsWith("https://");
  }

  // sourceHostname pulls the hostname out of a URL source for the
  // spinner label. Returns "" for non-URLs; URL-parse errors fall
  // back to the full trimmed source so the user still sees what
  // they pasted.
  function sourceHostname(source: string): string {
    const s = source.trim();
    if (!looksLikeURL(s)) return "";
    try {
      return new URL(s).hostname.replace(/^www\./, "");
    } catch {
      return s;
    }
  }

  // clearDeckFeedback resets every banner for a single game — called
  // before firing a new upload so stale success/error state doesn't
  // linger alongside fresh output.
  function clearDeckFeedback(id: string): void {
    delete deckErrors[id];
    delete deckViolations[id];
    delete deckWarnings[id];
    delete deckSuccess[id];
  }

  async function refresh(): Promise<void> {
    try {
      games = await listGames();
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "list failed";
    }
  }

  async function onCreate(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!newName.trim()) return;
    busy = true;
    error = "";
    try {
      const meta = await createGame(newName.trim());
      if (meta.invite_token) recentInvites.set(meta.id, meta.invite_token);
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

  function copyInvite(id: string): void {
    const token = recentInvites.get(id);
    if (!token) {
      error = "no invite token cached for this game — re-open as admin to recover";
      return;
    }
    const url = inviteURL(id, token);
    void navigator.clipboard.writeText(url);
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

  async function onUploadDeck(g: GameMeta): Promise<void> {
    const s = $session;
    if (!s?.playerID || s.gameID !== g.id) return;
    const source = deckSources[g.id] ?? "";
    if (!source.trim()) {
      clearDeckFeedback(g.id);
      deckErrors[g.id] = "paste a decklist first";
      return;
    }
    deckBusy = g.id;
    clearDeckFeedback(g.id);
    try {
      const res = await uploadDeck(g.id, s.playerID, source);
      deckSuccess[g.id] =
        `uploaded ${res.deck_name || "deck"}: ${res.card_count} cards, commander: ${res.commanders.join(", ")}`;
      if (res.warnings && res.warnings.length > 0) deckWarnings[g.id] = res.warnings;
      deckSources[g.id] = "";
      await refresh();
    } catch (err) {
      if (err instanceof LobbyApiError) {
        deckErrors[g.id] = err.message;
        if (err.violations && err.violations.length > 0) deckViolations[g.id] = err.violations;
        if (err.warnings && err.warnings.length > 0) deckWarnings[g.id] = err.warnings;
      } else {
        deckErrors[g.id] = "upload failed";
      }
    } finally {
      deckBusy = "";
    }
  }

  // canStart returns true when Start would succeed: at least 2 seats
  // and every seat has uploaded a real deck.
  function canStart(g: GameMeta): boolean {
    return g.state === "lobby" && g.players.length >= 2 && g.players.every((p) => p.deck_uploaded);
  }

  onMount(() => {
    void refresh();
  });
</script>

<section>
  <header>
    <h1>cmd_and_ctrl · lobby</h1>
    <p>
      logged in as <strong>{$session?.principal.role}</strong>
      {#if $session?.principal.name}
        ({$session.principal.name})
      {/if}
      <button class="linkish" onclick={logout}>log out</button>
    </p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if $session?.principal.role === "admin"}
    <h2>create game</h2>
    <form onsubmit={onCreate}>
      <input type="text" placeholder="game name" bind:value={newName} />
      <button type="submit" disabled={busy || !newName.trim()}>create</button>
    </form>
  {/if}

  <h2>games</h2>
  {#if games.length === 0}
    <p class="muted">no games yet.</p>
  {:else}
    <ul class="games">
      {#each games as g (g.id)}
        {@const seat = mySeat(g)}
        <li>
          <div class="row">
            <div>
              <strong>{g.name}</strong>
              <span class="muted">· {g.state} · {g.players.length} player(s)</span>
            </div>
            <div class="row-actions">
              {#if recentInvites.has(g.id)}
                <button onclick={() => copyInvite(g.id)}>copy invite</button>
              {/if}
              {#if canStart(g)}
                <button onclick={() => onStart(g.id)}>start</button>
              {:else if g.state === "lobby" && g.players.length >= 2}
                <button disabled title="waiting for all seats to upload a deck">start</button>
              {/if}
              <button onclick={() => openGame(g.id)}>open</button>
            </div>
          </div>
          <ul class="seats">
            {#each g.players as p (p.player_id)}
              <li>
                seat {p.seat}: {p.name}
                {#if p.deck_uploaded}
                  <span class="badge-ok">✓ {p.deck_name || "deck ready"}</span>
                {:else}
                  <span class="badge-pending">deck pending</span>
                {/if}
              </li>
            {/each}
          </ul>

          {#if seat && g.state === "lobby"}
            <details class="deck-upload" open={!seat.deck_uploaded}>
              <summary>
                {seat.deck_uploaded ? "replace your deck" : "upload your deck"}
              </summary>
              <p class="muted">
                Paste a Moxfield or Archidekt deck URL, a Moxfield JSON export, or a plain-text
                decklist. The server validates against Commander rules (100-card singleton, color
                identity, format legality).
              </p>
              <textarea
                rows="8"
                placeholder={"https://moxfield.com/decks/abc123\n\n— or —\n\nCommander:\n1 Atraxa, Praetors' Voice\n\nMainboard:\n1 Sol Ring\n..."}
                bind:value={deckSources[g.id]}
              ></textarea>
              <div class="row-actions">
                <button onclick={() => onUploadDeck(g)} disabled={deckBusy === g.id}>
                  {#if deckBusy === g.id}
                    {#if looksLikeURL(deckSources[g.id] ?? "")}
                      fetching from {sourceHostname(deckSources[g.id] ?? "")}…
                    {:else}
                      uploading…
                    {/if}
                  {:else if looksLikeURL(deckSources[g.id] ?? "")}
                    import deck
                  {:else}
                    upload deck
                  {/if}
                </button>
              </div>
              {#if deckErrors[g.id]}
                <pre class="error deck-error">{deckErrors[g.id]}</pre>
              {/if}
              {#if deckViolations[g.id]?.length}
                <ul class="violations">
                  {#each deckViolations[g.id] as v (v.code + (v.card ?? "") + v.message)}
                    <li><span class="code">{v.code}</span> · {violationLabel(v)}</li>
                  {/each}
                </ul>
              {/if}
              {#if deckWarnings[g.id]?.length}
                <ul class="warnings">
                  {#each deckWarnings[g.id] as v (v.code + (v.card ?? "") + v.message)}
                    <li><span class="code">{v.code}</span> · {violationLabel(v)}</li>
                  {/each}
                </ul>
              {/if}
              {#if deckSuccess[g.id]}
                <p class="success">{deckSuccess[g.id]}</p>
              {/if}
            </details>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  <p>
    <button onclick={refresh}>refresh</button>
  </p>
</section>

<style>
  section {
    max-width: 720px;
    margin: 2rem auto;
    padding: 1.5rem;
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
  }
  .games {
    list-style: none;
    padding: 0;
  }
  .games > li {
    border: 1px solid #ddd;
    padding: 0.75rem;
    border-radius: 4px;
    margin-bottom: 0.5rem;
  }
  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .row-actions {
    display: flex;
    gap: 0.5rem;
  }
  .seats {
    margin: 0.5rem 0 0 1rem;
    color: #555;
    font-size: 0.9em;
  }
  .muted {
    color: #666;
  }
  .error {
    color: #c00;
  }
  .success {
    color: #060;
    margin-top: 0.5rem;
  }
  .deck-error {
    white-space: pre-wrap;
    margin-top: 0.5rem;
    background: #fee;
    padding: 0.5rem;
    border-radius: 3px;
  }
  .violations,
  .warnings {
    margin-top: 0.5rem;
    padding: 0.5rem 0.75rem 0.5rem 1.5rem;
    border-radius: 3px;
    font-size: 0.85em;
  }
  .violations {
    background: #fee;
    color: #900;
  }
  .warnings {
    background: #ffb;
    color: #660;
  }
  .violations li,
  .warnings li {
    margin: 0.15rem 0;
  }
  .code {
    font-family: monospace;
    font-size: 0.9em;
    opacity: 0.75;
  }
  .badge-ok {
    color: #060;
    margin-left: 0.5rem;
  }
  .badge-pending {
    color: #a60;
    margin-left: 0.5rem;
  }
  .deck-upload {
    margin-top: 0.75rem;
    padding: 0.5rem;
    background: #f7f7f7;
    border-radius: 3px;
  }
  .deck-upload summary {
    cursor: pointer;
    font-weight: 600;
  }
  textarea {
    width: 100%;
    font-family: monospace;
    font-size: 0.9em;
    padding: 0.5rem;
    margin-top: 0.5rem;
    box-sizing: border-box;
  }
  form {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }
  input {
    flex: 1;
    padding: 0.5rem;
  }
  .linkish {
    background: none;
    border: none;
    color: #06c;
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
    margin-left: 0.5rem;
  }
</style>
