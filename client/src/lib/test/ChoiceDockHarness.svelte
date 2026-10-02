<script lang="ts">
  // ChoiceDockHarness mounts ChoicePromptModal beside a real ActionDock
  // (ADR 0111 PR 5), for render tests of one prompt kind that do not
  // need the whole Game route. Since PR 5 the small kinds open a dock
  // request instead of a modal, so a test of one of them reads the
  // dock; since PR 6 every other kind is a sheet in the same dock.
  import type { ActionType, GameView } from "../protocol";
  import type { ServerErrorLike } from "../choiceRejection";
  import ChoicePromptModal from "../components/board/ChoicePromptModal.svelte";
  import ActionDock from "../components/board/ActionDock.svelte";

  const {
    snap,
    viewerID,
    sendAction,
    lastError = null,
  }: {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: ActionType, params?: unknown, player?: string) => void;
    lastError?: ServerErrorLike | null;
  } = $props();
</script>

<ChoicePromptModal {snap} {viewerID} {sendAction} {lastError} />
<ActionDock
  view={snap}
  viewerHasPriority={false}
  viewerIsActive={false}
  autopassEnabled={false}
  onPassPriority={() => {}}
  onPassTurn={() => {}}
  onToggleAutopass={() => {}}
/>
