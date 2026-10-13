<script lang="ts">
  // The table's Discord voice room (ADR 0159 §6), as a section of the
  // table settings modal. Everyone sees what the room is doing and the
  // one thing they can do about it; the host (or an admin) also gets the
  // switch. The server enforces all of it: a disabled control is only so
  // the click never happens.

  import type { VoiceView } from "../protocol";
  import { voiceStatusLine, voiceYouLine } from "../voiceRoom";

  interface Props {
    voice: VoiceView;
    // Host or admin (canManageTable).
    canManage: boolean;
    // The game has ended: the room can be turned off, not on.
    ended: boolean;
    // The community server's invite, for a player who is not in it yet.
    inviteURL: string | null;
    ontoggle: (enabled: boolean) => void;
    onrefresh: () => void;
    busy?: boolean;
    error?: string | null;
  }

  const {
    voice,
    canManage,
    ended,
    inviteURL,
    ontoggle,
    onrefresh,
    busy = false,
    error = null,
  }: Props = $props();

  const you = $derived(voice.you ?? "spectator");
  const youLine = $derived(voiceYouLine(voice));
  const canReopen = $derived(
    voice.enabled && voice.status === "idle" && you !== "spectator" && you !== "no_discord",
  );
</script>

<section class="voice-room" aria-labelledby="voice-room-title">
  <h3 id="voice-room-title">Voice room</h3>
  <p class="voice-status" data-testid="voice-status">{voiceStatusLine(voice)}</p>
  {#if youLine}
    <p class="voice-you" data-testid="voice-you">{youLine}</p>
  {/if}

  <div class="voice-actions">
    {#if voice.url && you === "granted" && voice.status === "ready"}
      <a class="btn primary" href={voice.url} target="_blank" rel="noopener noreferrer"
        >Open in Discord</a
      >
    {/if}
    {#if you === "not_in_server"}
      {#if inviteURL}
        <a class="btn" href={inviteURL} target="_blank" rel="noopener noreferrer"
          >Join the Discord</a
        >
      {/if}
      <button type="button" disabled={busy} onclick={onrefresh}>I've joined — check again</button>
    {/if}
    {#if canReopen}
      <button type="button" disabled={busy} onclick={onrefresh}>Reopen voice room</button>
    {/if}
  </div>

  {#if canManage}
    <label class="voice-toggle">
      <input
        type="checkbox"
        checked={voice.enabled}
        disabled={busy || (ended && !voice.enabled)}
        onchange={(e) => ontoggle((e.currentTarget as HTMLInputElement).checked)}
      />
      <span>Give this table a voice room</span>
    </label>
  {/if}
  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
</section>

<style>
  .voice-room {
    margin-top: 1rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border, rgba(127, 127, 127, 0.3));
  }
  h3 {
    margin: 0 0 0.35rem;
    font-size: 0.95rem;
  }
  .voice-status,
  .voice-you {
    margin: 0.25rem 0;
  }
  .voice-you {
    opacity: 0.85;
  }
  .voice-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin: 0.5rem 0;
  }
  .voice-actions:empty {
    display: none;
  }
  .voice-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }
</style>
