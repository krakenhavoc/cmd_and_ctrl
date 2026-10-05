<script lang="ts">
  // TipCard — the tip on screen, wired to the runtime: a HintCard whose
  // buttons dismiss through lib/hints/runtime.ts. HintLayer draws one
  // over the page; HintSlot draws one inside the Settings dialog.

  import {
    dismissTip,
    rememberReturnFocus,
    TIP_BODY_ID,
    type ActiveTip,
  } from "../../hints/runtime";
  import HintCard from "./HintCard.svelte";

  interface Props {
    tip: ActiveTip;
  }

  const { tip }: Props = $props();
</script>

<HintCard
  title={tip.title}
  body={tip.body}
  bodyId={TIP_BODY_ID}
  placement={tip.placement}
  ring={tip.ring}
  stripBottom={tip.stripBottom}
  reduceMotion={tip.reduceMotion}
  action={tip.hint.action}
  onGotIt={() => dismissTip("got-it")}
  onHideTips={() => dismissTip("hide")}
  onAction={() => dismissTip("got-it")}
  onFocusFrom={rememberReturnFocus}
/>
