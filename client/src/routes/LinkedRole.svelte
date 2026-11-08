<script lang="ts">
  // LinkedRole — where Discord Linked Roles lands, #/linked-role (ADR 0163 §3).
  //
  // Discord sends a person to the server's verification URL
  // (/auth/discord/linked-role); after Discord's consent screen the server
  // sends them here with how it ended. Linking never signs anyone in, so
  // the page is public and holds no account data.
  //
  //   linked       their numbers are on their Discord profile now
  //   signin       no CMD & CTRL account for that Discord account: nothing
  //                was stored; sign in first, then link again
  //   unavailable  this server does not offer Linked Roles
  //   failed       Discord said no, or did not answer; nothing was kept
  import Page from "../lib/components/Page.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import { discordLoginHref } from "../lib/api";
  import type { LinkedRoleStatus } from "../lib/router";

  interface Props {
    status: LinkedRoleStatus;
  }
  const { status }: Props = $props();

  const LINK_HREF = "/auth/discord/linked-role";

  const copy: Record<LinkedRoleStatus, { eyebrow: string; title: string; lede: string }> = {
    linked: {
      eyebrow: "Linked",
      title: "Linked. You can go back to Discord.",
      lede: "Your Discord profile now shows your CMD & CTRL games played, games won and the date you joined. They update when a game you sat in ends, and the server's roles can use them.",
    },
    signin: {
      eyebrow: "Almost there",
      title: "Sign in to CMD & CTRL first, then link again",
      lede: "There is no CMD & CTRL account for this Discord account yet, so nothing was linked and nothing was kept. Sign in once, then come back to the role in Discord and link again.",
    },
    unavailable: {
      eyebrow: "Linked Roles",
      title: "Linking isn't available on this server",
      lede: "This CMD & CTRL server doesn't offer Discord Linked Roles. Nothing was linked and nothing was kept.",
    },
    failed: {
      eyebrow: "Linked Roles",
      title: "That didn't link — try again",
      lede: "Discord didn't finish the link, or didn't answer. Nothing was kept. Trying again from Discord usually works.",
    },
  };
  const c = $derived(copy[status]);
</script>

{#snippet actions()}
  {#if status === "signin"}
    <a class="btn discord" href={discordLoginHref()}><Icon name="discord" size={15} /> Sign in</a>
  {:else if status === "failed"}
    <a class="btn primary" href={LINK_HREF}>Link again</a>
  {/if}
{/snippet}

<Page
  eyebrow={c.eyebrow}
  title={c.title}
  lede={c.lede}
  actions={status === "signin" || status === "failed" ? actions : undefined}
  width="prose"
>
  <p class="note">
    {#if status === "linked"}
      To stop showing them, remove the CMD & CTRL connection in Discord under User Settings ›
      Connections. Linking grants nothing on this site.
    {:else}
      Linked Roles only shows numbers on your Discord profile; it never grants anything on this
      site. See the <a href="#/privacy">privacy notice</a> for what is shared.
    {/if}
  </p>
</Page>

<style>
  .note {
    margin: 0;
    font-size: 13.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
</style>
