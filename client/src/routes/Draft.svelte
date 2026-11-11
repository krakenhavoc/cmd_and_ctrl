<script lang="ts">
  // Draft — a booster draft (ADR 0222 §5, #5438).
  //
  //   #/draft                       the create form (DraftCreate)
  //   #/draft/<id>?invite=<code>    the invite page (DraftJoin), for a
  //                                 browser that holds no seat there
  //   #/draft/<id>                  the draft itself (DraftRoom): its
  //                                 lobby, the picks and the finished pool
  //
  // Public, like the practice table (lib/signedInHome.ts): a visitor
  // drafts against bots with no sign-in.
  //
  // The seat is a token kept in localStorage under the draft's id
  // (lib/draft.ts), not the session cookie, so a reload or a return visit
  // comes back to the same seat and the same picks. When the server says
  // the seat is gone the token is forgotten: a link that still carries
  // its invite offers the join page again, and any other shows a plain
  // message.
  import {
    DRAFT_NEW_HASH,
    rememberDraftSeat,
    storedDraftSeat,
    type DraftSession,
  } from "../lib/draft";
  import { L } from "../lib/labels";
  import DraftCreate from "../lib/components/draft/DraftCreate.svelte";
  import DraftJoin from "../lib/components/draft/DraftJoin.svelte";
  import DraftRoom from "../lib/components/draft/DraftRoom.svelte";
  import Page from "../lib/components/Page.svelte";

  interface Props {
    /** The draft's id; absent on the create form. */
    id?: string;
    /** ?invite=<code>, on a link a host shared. */
    invite?: string;
    /** ?set=<code>, on the create form. */
    set?: string;
  }
  const { id, invite, set }: Props = $props();

  // App.svelte keys this route on the id, so the seat is read once.
  // svelte-ignore state_referenced_locally
  let token = $state(id ? (storedDraftSeat(id)?.token ?? "") : "");

  function joined(s: DraftSession): void {
    if (!id) return;
    rememberDraftSeat(id, s);
    token = s.token;
  }
</script>

{#if !id}
  <DraftCreate {set} />
{:else if token}
  {#key token}
    <DraftRoom {id} {token} ongone={() => (token = "")} />
  {/key}
{:else if invite}
  <DraftJoin {id} {invite} onjoined={joined} />
{:else}
  <Page eyebrow="Draft" title={L.draftGone} width="prose">
    <div class="card">
      <p>
        This browser holds no seat in that draft. The draft may have closed, or your seat in it may
        have ended. If someone invited you, open the invite link they sent.
      </p>
      <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewDraft} →</a>
    </div>
  </Page>
{/if}

<style>
  p {
    margin: 0 0 12px;
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
</style>
