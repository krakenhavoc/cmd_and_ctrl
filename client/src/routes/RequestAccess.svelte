<script lang="ts">
  // RequestAccess — the gate page, #/request-access (ADR 0156 §5).
  //
  // The Discord callback sends a sign-in here when the role gate turned
  // it away: no session was issued and nothing was stored. The page says
  // why, in the steps the C2 server's onboarding actually takes, and
  // offers the two ways in that need no account: a guest practice game,
  // and a host's invite link.
  //
  // `retry` is ?retry=1: Discord did not answer the check, so the page
  // leads with that and with Check again.
  //
  // The role and channel names are copy, not config: renaming them on
  // Discord means changing them here. The Join button hides when the
  // server names no invite (GET /community).
  import Page from "../lib/components/Page.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import { discordLoginHref } from "../lib/api";
  import { community } from "../lib/community";
  import { L } from "../lib/labels";

  interface Props {
    retry?: boolean;
  }
  const { retry = false }: Props = $props();

  const invite = $derived($community.invite_url);
  // Discord remembers the consent, so this is a redirect, not a prompt.
  const checkHref = discordLoginHref();
</script>

{#snippet checkAgain()}
  <a class="btn primary" href={checkHref}>Check again</a>
{/snippet}

<Page
  eyebrow={retry ? "Sign-in" : "Almost there"}
  title={retry ? "Discord didn't answer — try again" : "Ask for an account on our Discord"}
  lede={retry
    ? "We couldn't check your account with Discord just now. Checking again usually works; nothing was saved in the meantime."
    : "Accounts are for members of the C2 Discord who have the Elder Dragon Highlander role. Here's how to get it."}
  actions={retry ? checkAgain : undefined}
  width="prose"
>
  <ol class="steps" aria-label="how to get an account">
    <li class="step">
      <span class="n" aria-hidden="true">1</span>
      <div class="body">
        <h2>Join the C2 Discord</h2>
        <p>Everyone who joins is a Planeswalker: you can chat, find a pod and play as a guest.</p>
      </div>
      {#if invite}
        <a class="btn discord sm" href={invite} target="_blank" rel="noopener noreferrer"
          ><Icon name="discord" size={15} /> Join</a
        >
      {/if}
    </li>
    <li class="step">
      <span class="n" aria-hidden="true">2</span>
      <div class="body">
        <h2>Answer yes to “request account access”</h2>
        <p>
          It's a question in the server's onboarding. Already joined and answered no? Change it
          under the server's <strong>Channels &amp; Roles</strong>.
        </p>
      </div>
    </li>
    <li class="step">
      <span class="n" aria-hidden="true">3</span>
      <div class="body">
        <h2>Ask in the access-request channel</h2>
        <p>Your answer opens it. A mod gives you the Elder Dragon Highlander role.</p>
      </div>
    </li>
    <li class="step">
      <span class="n hot" aria-hidden="true">4</span>
      <div class="body">
        <h2>Check again</h2>
        <p>Signs you in with Discord once more. It remembers you, so there's nothing to approve.</p>
      </div>
      {@render checkAgain()}
    </li>
  </ol>

  <div class="meanwhile">
    <a class="btn ghost" href="#/practice"><Icon name="robot" size={15} /> {L.tryPracticeGame}</a>
    <p>
      In the meantime, a practice game needs no account, and an invite link from a host still works:
      you sit at that table as a guest.
    </p>
  </div>
</Page>

<style>
  .steps {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .step {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 18px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--surface);
  }
  .n {
    flex: none;
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border-radius: 9px;
    background: var(--surface-raised);
    color: var(--fg);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 14px;
  }
  .n.hot {
    color: var(--accent);
    background: var(--accent-soft);
  }
  .body {
    flex: 1;
    min-width: 0;
  }
  .step h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 15.5px;
    font-weight: 700;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--fg);
  }
  .step p {
    margin: 4px 0 0;
    font-size: 13.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .step p strong {
    color: var(--fg);
  }
  .step .btn {
    flex: none;
  }
  .meanwhile {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 18px;
    margin-top: 28px;
    padding-top: 22px;
    border-top: 1px solid var(--border);
  }
  .meanwhile p {
    flex: 1 1 30ch;
    margin: 0;
    font-size: 13.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
</style>
