// @vitest-environment jsdom
//
// opensFlashCost.render.test.ts — #5363 (CR 601.3c): an optional cost
// marked `opens_flash` is the only way some claims can be cast right now,
// and the server says which (`opens_flash_required`). Choosing such a
// claim ticks the cost; while it is unticked the picker says why and Cast
// waits. A claim that does not need it leaves the cost unticked.

import { describe, it, expect, afterEach } from "vitest";

import AlternativeCostModal from "./components/board/AlternativeCostModal.svelte";
import type { CardView } from "./protocol";
import { L } from "./labels";
import { opensFlashCostToPreselect, opensFlashRequired } from "./targeting";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { barPrimary } from "./test/dockView";

afterEach(cleanup);

// Asinine Antics: "You may cast this spell as though it had flash if you
// pay {2} more to cast it." `printedNeeds` is the printed claim's flag;
// `free` adds Omniscience's offer, with its own flag.
function antics(printedNeeds: boolean, free?: { needs: boolean }): CardView {
  return {
    instance_id: "antics",
    name: "Asinine Antics",
    owner: "me",
    controller: "me",
    mana_cost: "{2}{U}{U}",
    opens_flash_required: printedNeeds || undefined,
    optional_costs: [
      {
        index: 0,
        key: "flash-for-more",
        label: "Pay {2} more",
        mana_cost: "{2}",
        opens_flash: true,
      },
    ],
    alternative_costs: free
      ? [
          {
            key: "granted-free",
            label: "Without paying its mana cost",
            opens_flash_required: free.needs || undefined,
          },
        ]
      : undefined,
  } as CardView;
}

function mount(c: CardView) {
  const confirmed: Array<{ key: string | undefined; optional: number[] }> = [];
  const view = render(
    DockHarness as never,
    {
      component: AlternativeCostModal,
      props: {
        card: c,
        onConfirm: (key: string | undefined, optional: number[]) =>
          confirmed.push({ key, optional }),
        onCancel: () => {},
      },
    } as never,
  );
  return { container: view.container, confirmed };
}

const flashRow = (c: HTMLElement): HTMLButtonElement => {
  const row = [...c.querySelectorAll<HTMLButtonElement>(".dock-sheet .prompt-opt")].find((b) =>
    b.textContent?.includes("Pay {2} more"),
  );
  if (!row) throw new Error("no Pay {2} more row");
  return row;
};
const claimRow = (c: HTMLElement, name: string): HTMLButtonElement => {
  const row = [...c.querySelectorAll<HTMLButtonElement>(".dock-sheet .prompt-opt")].find(
    (b) => b.querySelector(".name")?.textContent?.trim() === name,
  );
  if (!row) throw new Error(`no ${name} row`);
  return row;
};
const hint = (c: HTMLElement): string | undefined =>
  c.querySelector<HTMLElement>(".flash-hint")?.textContent ?? undefined;

describe("an opens_flash cost the cast needs right now", () => {
  it("is ticked when the printed cost needs it, and the cast sends it", () => {
    const { container, confirmed } = mount(antics(true));
    expect(flashRow(container).getAttribute("aria-pressed")).toBe("true");
    expect(flashRow(container).textContent).toContain(L.opensFlashNeededNow);
    expect(hint(container)).toBeUndefined();
    click(barPrimary(container)!);
    expect(confirmed).toEqual([{ key: undefined, optional: [0] }]);
  });

  it("says why and holds Cast while it is unticked", () => {
    const { container, confirmed } = mount(antics(true));
    click(flashRow(container));
    expect(flashRow(container).getAttribute("aria-pressed")).toBe("false");
    expect(hint(container)).toContain("You can cast this now only by also paying “Pay {2} more”");
    const cast = barPrimary(container)! as HTMLButtonElement;
    expect(cast.disabled).toBe(true);
    click(cast);
    expect(confirmed).toEqual([]);
  });

  it("is left alone when no claim needs it", () => {
    const { container, confirmed } = mount(antics(false));
    expect(flashRow(container).getAttribute("aria-pressed")).toBe("false");
    expect(flashRow(container).textContent).not.toContain(L.opensFlashNeededNow);
    expect(hint(container)).toBeUndefined();
    click(barPrimary(container)!);
    expect(confirmed).toEqual([{ key: undefined, optional: [] }]);
  });

  it("is ticked when the player picks an offer that needs it", () => {
    const { container, confirmed } = mount(antics(false, { needs: true }));
    expect(flashRow(container).getAttribute("aria-pressed")).toBe("false");
    click(claimRow(container, "Without paying its mana cost"));
    expect(flashRow(container).getAttribute("aria-pressed")).toBe("true");
    click(barPrimary(container)!);
    expect(confirmed).toEqual([{ key: "granted-free", optional: [0] }]);
  });
});

describe("opensFlashCostToPreselect", () => {
  const payable = () => true;
  it("names the cost only for a claim that needs it and pays none yet", () => {
    const c = antics(false, { needs: true });
    expect(opensFlashRequired(c, undefined)).toBe(false);
    expect(opensFlashRequired(c, "granted-free")).toBe(true);
    expect(opensFlashCostToPreselect(c, undefined, new Map(), payable)).toBeUndefined();
    expect(opensFlashCostToPreselect(c, "granted-free", new Map(), payable)).toBe(0);
    expect(
      opensFlashCostToPreselect(c, "granted-free", new Map([[0, 1]]), payable),
    ).toBeUndefined();
    expect(opensFlashCostToPreselect(c, "granted-free", new Map(), () => false)).toBeUndefined();
  });
});
