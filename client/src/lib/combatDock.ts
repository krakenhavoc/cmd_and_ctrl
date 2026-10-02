// combatDock.ts — combat's requests for the action dock (ADR 0111
// Delivery PR 3). Pure builders: Game.svelte owns the state and the
// handlers, and these turn them into the DockRequest the dock draws.
// They replace the attention strip's attack, block and combat-hint
// clusters and the attack-tax / attack-limit refusal toasts.
//
// What each one is (ADR 0111 §1, "Steps"):
//
//   - The attack row (rank `step`): "N ready" and Attack all, one
//     button per opponent at a wider table, Choose attackers… where a
//     tax or a limit applies. It does not take the action bar: in this
//     engine you keep declaring while you hold priority, and `next` is
//     what ends the declaration. The declaration's Undo is not here any
//     more; it is the dock's one Undo, in the toggles row.
//   - The block request (rank `blocks`): "No blocks" or "Done blocking"
//     as the action bar's primary.
//   - The combat selection (rank `flow`): the hint as the question line
//     and Cancel as the secondary. No primary: the click on the board
//     commits.
//
// The names are the strip's (ADR 0111 §10): "declare attackers" and
// "declare blockers" groups, "Attack <name> with all N creatures",
// "Choose attackers…", "No blocks", "Done blocking", "Cancel".

import {
  attackAllLabel,
  attackAllTaxLabel,
  attackTaxOn,
  eligibleAt,
  offersAttackPicker,
  seatLabel,
  type AttackAllPlan,
  type BulkAttackRefusal,
} from "./attackAll";
import type { DockAction, DockRefusal, DockRequest } from "./dock";
import type { GameView, PlayerView } from "./protocol";

export interface AttackRowInput {
  view: GameView;
  plan: AttackAllPlan;
  // canDeclareAttackers && something eligible && someone to attack.
  ready: boolean;
  // blockedSummary(plan.blocked).
  blockedHint: string;
  // The bound attack-all chord (offered on the duel button only: the
  // shortcut refuses to guess between opponents).
  attackAllChord?: string;
  seatColor: (seat: number) => string;
  onAttackAll: (defenderSeatID: string) => void;
  onChooseAttackers: (defenderSeatID: string) => void;
  // The bulk refusal the dock answers, with its server error.
  refusal?: {
    kind: BulkAttackRefusal["kind"];
    message: string;
    reason?: string;
    missing?: string[];
    // attackLimitOn for the refused seat: null unknown, 0 used up.
    limitRoom: number | null;
    onChoose: () => void;
    onDismiss: () => void;
  } | null;
}

// ADR 0080 (#1063): the button's tooltip names the CR 508.1a price as
// well as the count, so the cost of a wide swing under Propaganda is
// legible before the click rather than arriving as a refusal. The price
// is the server's; nothing here derives it (#429).
function attackAllTitle(view: GameView, plan: AttackAllPlan, opp: PlayerView): string {
  return [attackAllLabel(plan, opp), attackAllTaxLabel(view, plan, opp.id)]
    .filter(Boolean)
    .join(" · ");
}

function attackQuestion(input: AttackRowInput): { question: string; detail: string } {
  const { plan, ready, blockedHint } = input;
  if (ready) {
    const detail = [
      plan.declared.length > 0 ? `${plan.declared.length} already declared` : "",
      blockedHint ? `can't: ${blockedHint}` : "",
    ]
      .filter(Boolean)
      .join(" · ");
    return { question: `${plan.eligible.length} ready to attack`, detail };
  }
  return {
    question: `${plan.declared.length} declared`,
    detail: blockedHint ? `${blockedHint} can't attack` : "",
  };
}

function attackRow(input: AttackRowInput): { lead?: string; row: DockAction[] } {
  const { view, plan } = input;
  if (!input.ready) return { row: [] };
  if (plan.defenders.length === 1) {
    const opp = plan.defenders[0]!;
    const row: DockAction[] = [
      {
        id: `attack-all:${opp.id}`,
        label: attackAllLabel(plan, opp),
        note: attackAllTaxLabel(view, plan, opp.id) || undefined,
        title: attackAllTitle(view, plan, opp),
        chord: input.attackAllChord,
        disabled: eligibleAt(plan, opp.id).length === 0,
        emphasis: true,
        onPress: () => input.onAttackAll(opp.id),
      },
    ];
    if (offersAttackPicker(view, plan, opp.id)) {
      // #1162 / #1533: the seat can afford, or may send, only SOME of a
      // wide swing — offered up front, not only after a refusal.
      row.push({
        id: `choose:${opp.id}`,
        label: "Choose attackers…",
        title: attackTaxOn(view, opp.id)
          ? "pick which attackers to send, and lock a land against the auto-tapper"
          : "pick which attackers to send — an effect limits how many can attack",
        onPress: () => input.onChooseAttackers(opp.id),
      });
    }
    return { row };
  }
  // Several opponents: one button per seat rather than a bare "attack
  // all", so the control always says who gets hit. Nothing spreads an
  // attack.
  const row: DockAction[] = [];
  for (const opp of plan.defenders) {
    row.push({
      id: `attack-all:${opp.id}`,
      label: seatLabel(opp),
      note: attackTaxOn(view, opp.id) || undefined,
      seatColor: input.seatColor(opp.seat),
      title: attackAllTitle(view, plan, opp),
      disabled: eligibleAt(plan, opp.id).length === 0,
      onPress: () => input.onAttackAll(opp.id),
    });
    if (offersAttackPicker(view, plan, opp.id)) {
      row.push({
        id: `choose:${opp.id}`,
        label: "",
        icon: "more",
        ariaLabel: `Choose attackers against ${seatLabel(opp)}`,
        title: attackTaxOn(view, opp.id)
          ? `pick which attackers to send at ${seatLabel(opp)}, and lock a land against the auto-tapper`
          : `pick which attackers to send at ${seatLabel(opp)} — an effect limits how many can attack`,
        onPress: () => input.onChooseAttackers(opp.id),
      });
    }
  }
  return { lead: "Attack all →", row };
}

// attackRefusal is the attack-tax (#1162) or attack-limit (#1533)
// refusal of the last "attack with all", with the picker it offers.
function attackRefusal(r: NonNullable<AttackRowInput["refusal"]>): DockRefusal {
  if (r.kind === "tax") {
    return {
      tag: "attack tax",
      text: r.reason
        ? `Attacking with all of them costs ${r.reason} and you can't pay it`
        : "You can't pay to attack with all of them",
      detail: r.missing && r.missing.length > 0 ? `missing ${r.missing.join(" ")}` : undefined,
      actions: [{ id: "refusal-choose", label: "Choose attackers…", onPress: r.onChoose }],
      onDismiss: r.onDismiss,
    };
  }
  return {
    tag: "attack limit",
    text: r.message,
    detail: r.limitRoom === 0 ? "no more creatures can attack this combat" : undefined,
    actions:
      r.limitRoom === 0
        ? []
        : [
            {
              id: "refusal-choose",
              label: r.limitRoom === null ? "Choose attackers…" : `Choose up to ${r.limitRoom}…`,
              onPress: r.onChoose,
            },
          ],
    onDismiss: r.onDismiss,
  };
}

export function attackRowRequest(input: AttackRowInput): DockRequest {
  const { question, detail } = attackQuestion(input);
  const { lead, row } = attackRow(input);
  return {
    rank: "step",
    label: "declare attackers",
    group: "declare attackers",
    tag: "attack",
    tone: "danger",
    question,
    detail: detail || undefined,
    rowLead: lead,
    row,
    refusal: input.refusal ? attackRefusal(input.refusal) : null,
  };
}

// #1279: the viewer is a defender whose declaration is still open.
// Whatever is staged is the declaration; nothing staged is "no blocks".
//
// Enter (ADR 0111 PR 4) presses Done blocking, which commits blocks the
// player already put down. It does not press No blocks: that declines a
// window that cannot be got back (#328), and since the owner's
// 2026-10-02 decision it also passes priority, so it takes a click (or
// Enter on the button itself once it has focus), never a stray Enter.
export function blockRequest(staged: number, onFinish: () => void): DockRequest {
  const done = staged > 0;
  return {
    rank: "blocks",
    label: "declare blockers",
    group: "declare blockers",
    tag: "block",
    tone: "plain",
    question:
      staged > 0
        ? `${staged} ${staged === 1 ? "blocker" : "blockers"} declared`
        : "Choose blockers, or declare none",
    primary: {
      id: "finish-blocks",
      label: done ? "Done blocking" : "No blocks",
      title: done
        ? "finish declaring blockers — the attacking player gets priority once every defender is done"
        : "declare no blockers — and pass priority, if you hold it",
      keyShortcuts: done ? "Enter" : undefined,
      cap: done ? "⏎" : undefined,
      onPress: onFinish,
    },
  };
}

export function combatSelectionRequest(
  kind: "attacker" | "blocker",
  cardName: string,
  onCancel: () => void,
): DockRequest {
  const attacking = kind === "attacker";
  return {
    rank: "flow",
    label: `${attacking ? "attacking" : "blocking"} with ${cardName}`,
    tag: attacking ? "attack" : "block",
    tone: "danger",
    live: true,
    question: attacking
      ? `Attacking with ${cardName} — click an opponent's seat to commit, or the creature again to cancel.`
      : `Blocking with ${cardName} — click an incoming attacker to commit, or the creature again to cancel.`,
    primary: null,
    secondary: [
      {
        id: "cancel",
        label: "Cancel",
        cap: "Esc",
        keyShortcuts: "Escape",
        title: "cancel the selection",
        onPress: onCancel,
      },
    ],
  };
}
