// shopLinksStore.ts — which shops a viewer sees links for (ADR 0179).
//
// The operator's configured shops (GET /config, env.ts), less the
// viewer's own "Show shop links" setting (settings.display.shopLinks,
// synced to a signed-in account, kept in this browser for a guest).
// Empty means render nothing: no links and no disclosure.
//
// This store says nothing about WHERE links may go. The caller keeps
// the rule that none renders on the game table while a game is running;
// see ADR 0179 §3 for the placements.

import { type Readable } from "svelte/store";

import { appConfig } from "./env";
import { guardedDerived, guardedWritable } from "./guardedStore";
import { settings, updateSettings } from "./settings";
import type { Shop } from "./shopLinks";

/** The shops the operator configured, whatever the viewer chose. */
export const configuredShops: Readable<Shop[]> = guardedDerived(
  appConfig,
  (c) => c.shops,
  "configuredShops",
  [],
);

// The two inputs, kept together. Both parents live as long as the page,
// so these subscriptions are never dropped.
const inputs = guardedWritable<{ shops: Shop[]; shown: boolean }>(
  { shops: [], shown: true },
  "shopLinkInputs",
);
appConfig.subscribe((c) => inputs.update((v) => ({ ...v, shops: c.shops })));
settings.subscribe((s) => inputs.update((v) => ({ ...v, shown: s.display.shopLinks })));

/** The shops to render links for: none when hidden or unconfigured. */
export const visibleShops: Readable<Shop[]> = guardedDerived(
  inputs,
  ({ shops, shown }) => (shown ? shops : []),
  "visibleShops",
  [],
);

/** setShopLinksShown is the "Show shop links" setting's one writer. */
export function setShopLinksShown(shown: boolean): void {
  updateSettings("display", "shopLinks", shown);
}
