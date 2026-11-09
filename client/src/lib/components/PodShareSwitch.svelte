<script lang="ts">
  // PodShareSwitch — the one pod-sharing switch (ADR 0220 §7): "Share my
  // stats with my pods". It is stats_prefs.share_pods, one switch for
  // every pod the person is in, off until they turn it on; joining a pod
  // shares nothing. It lives in Settings → Stats and on a pod's page, and
  // both read the same store.
  import { L } from "../labels";
  import { SHARE_PODS_NOTE } from "../podStats";
  import { setSharePods, statsPrefs } from "../statsPrefs";

  interface Props {
    /** Called after the server accepts a change, so a page can reload. */
    onchange?: (on: boolean) => void;
  }
  const { onchange }: Props = $props();

  let busy = $state(false);
  let error = $state("");

  async function toggle(on: boolean): Promise<void> {
    busy = true;
    error = "";
    try {
      await setSharePods(on);
      onchange?.(on);
    } catch (e) {
      error = e instanceof Error ? e.message : "could not save; try again";
    } finally {
      busy = false;
    }
  }
</script>

<label class="switch-row" data-testid="share-pods">
  <input
    type="checkbox"
    checked={$statsPrefs.share_pods}
    disabled={busy}
    onchange={(e) => toggle(e.currentTarget.checked)}
  />
  <span>
    {L.sharePodsSwitch}
    <span class="help">{SHARE_PODS_NOTE}</span>
  </span>
</label>
{#if error}
  <p class="error" role="alert">{error}</p>
{/if}

<style>
  .switch-row {
    display: flex;
    gap: 0.6rem;
    align-items: flex-start;
    margin: 0.6rem 0;
  }
  .help {
    display: block;
    margin: 2px 0 0;
    color: var(--fg-muted);
    font-size: 11.5px;
    line-height: 1.45;
  }
  .error {
    color: var(--danger);
    margin: 0;
  }
</style>
