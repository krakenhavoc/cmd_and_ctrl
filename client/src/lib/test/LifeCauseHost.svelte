<script lang="ts">
  // LifeCauseHost — a render-test host that is Board.svelte as far as an
  // avatar can tell: it provides the life-cause contexts (lifeCause.ts) over
  // a view, and draws one seat's PlayerIdentity inside them. Test-only.
  import PlayerIdentity from "../components/board/PlayerIdentity.svelte";
  import { provideLifeCauses } from "../lifeCause";
  import type { GameView, PlayerView } from "../protocol";

  interface Props {
    view: GameView;
    seat: PlayerView;
  }
  const { view, seat }: Props = $props();

  provideLifeCauses(() => view);
</script>

<PlayerIdentity
  {seat}
  isSelf={false}
  isActive={false}
  hasPriority={false}
  attackTargetable={false}
  isMonarch={false}
  isInitiative={false}
  sendAction={() => {}}
/>
