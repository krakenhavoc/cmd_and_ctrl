<script lang="ts">
  // AccountLink: an account's avatar and name, opening its detail
  // (ADR 0124 §7, "a seat with an account opens that account").
  import { accountHash, avatarSrc, type AdminAccountRef } from "../../adminViews";
  import { session } from "../../session";

  interface Props {
    account: AdminAccountRef;
    /** the name to show when the account carries none */
    fallback?: string;
  }
  const { account, fallback = "account" }: Props = $props();

  const src = $derived(avatarSrc(account.avatar_url, $session?.token));
</script>

<a class="acct" href={accountHash(account.id)}>
  {#if src}<img class="avatar" {src} alt="" width="20" height="20" loading="lazy" />{/if}
  <span class="nm">{account.name || fallback}</span>
</a>

<style>
  .acct {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    max-width: 100%;
    color: var(--fg);
    font-weight: 600;
    text-decoration: none;
  }
  .acct:hover .nm {
    text-decoration: underline;
  }
  .nm {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .avatar {
    flex: none;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: var(--surface-raised);
  }
</style>
