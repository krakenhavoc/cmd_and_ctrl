<script lang="ts">
  // DockHarness mounts any picker beside a real ActionDock (ADR 0111
  // Delivery PR 6), for render tests of one picker that do not need the
  // whole Game route or Board. Since PR 6 every big picker is a sheet in
  // the dock (components/board/DockSheet.svelte): mounted alone it draws
  // nothing, so a test of one reads the dock.
  //
  //     render(DockHarness, { component: XCostModal, props: { card, … } })
  //
  // `props` is spread onto the picker, so `setProps({ props: {...} })`
  // re-renders it with new values.
  import type { Component } from "svelte";
  import type { GameView } from "../protocol";
  import ActionDock from "../components/board/ActionDock.svelte";
  import { dockTestView } from "./dockView";

  const {
    component,
    props = {},
    view = dockTestView(),
  }: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    component: Component<any>;
    props?: Record<string, unknown>;
    view?: GameView;
  } = $props();

  const Picker = $derived(component);
</script>

<Picker {...props} />
<ActionDock
  {view}
  viewerHasPriority={false}
  viewerIsActive={false}
  autopassEnabled={false}
  onPassPriority={() => {}}
  onPassTurn={() => {}}
  onToggleAutopass={() => {}}
/>
