import type { CardView, GameView } from "./protocol";

// attackTargets.ts — S27: an attacker may be declared against a
// player, a planeswalker or a battle (CR 506.2, 508.1d).
//
// The legal set is computed SERVER-side and rides
// `view.turn.attack_targets`; this module only turns it into menu
// rows. Deriving it here instead would mean the client and the server
// each holding an opinion about who protects which battle, and the
// one that matters is the server's.
//
// The set on the wire belongs to the ACTIVE player, because only the
// active player declares attackers. A menu opened on a card some
// other seat controls — an admin driving the table — therefore gets
// the permanent rows dropped rather than a set computed for the wrong
// seat. Seats still list, exactly as they did before S27.

// AttackTargetRow is one row of the "Declare attacker" submenu: the
// id to send as declare_attacker's `target`, and what to call it.
export interface AttackTargetRow {
  id: string;
  label: string;
  kind: "player" | "planeswalker" | "battle";
}

// permanentAttackTargets returns the planeswalker and battle rows for
// an attacker controlled by `controllerID`, or an empty list when the
// server has published no set for that seat.
//
// Player rows are deliberately NOT included: the existing menu builds
// those from `view.seats` and has done since S08, it carries the
// display-name fallback, and folding them in here would change what
// an unrelated row says.
export function permanentAttackTargets(view: GameView, controllerID: string): AttackTargetRow[] {
  const targets = view.turn?.attack_targets ?? [];
  if (targets.length === 0) return [];
  // Only the active player declares attackers, so a set published for
  // seat N is only meaningful for a card seat N controls.
  const active = view.seats[view.turn.active_seat];
  if (!active || active.id !== controllerID) return [];

  const byID = new Map<string, CardView>();
  for (const c of view.battlefield?.cards ?? []) byID.set(c.instance_id, c);

  const out: AttackTargetRow[] = [];
  for (const t of targets) {
    if (t.kind !== "planeswalker" && t.kind !== "battle") continue;
    const card = byID.get(t.id);
    out.push({
      id: t.id,
      label: card?.name || (t.kind === "battle" ? "battle" : "planeswalker"),
      kind: t.kind,
    });
  }
  return out;
}

// attackTargetHint is the secondary copy on a permanent row: whose
// planeswalker it is, or who is defending the battle. Both are the
// piece of information the attacking player actually needs — a
// planeswalker's controller blocks for it, and a battle's PROTECTOR
// does, which is not the same seat as its controller.
export function attackTargetHint(view: GameView, row: AttackTargetRow): string {
  const card = (view.battlefield?.cards ?? []).find((c) => c.instance_id === row.id);
  if (!card) return "";
  if (row.kind === "battle") {
    const seat = view.seats.find((s) => s.id === card.protector_player);
    const defense = card.defense ?? 0;
    const who = seat ? `defended by ${seat.display_name || seat.name}` : "undefended";
    return `${who} · ${defense} defense`;
  }
  const seat = view.seats.find((s) => s.id === card.controller);
  const loyalty = card.counters?.loyalty ?? 0;
  return seat ? `${seat.display_name || seat.name}'s · ${loyalty} loyalty` : `${loyalty} loyalty`;
}

// defendingPlayerOf is the seat defending against `attacker`'s attack
// — the only seat whose creatures may block it (CR 802.4a, #1339): the
// player attacked, the controller of the planeswalker attacked, or the
// PROTECTOR of the battle attacked. The server computes it and ships it
// as `defending_player`, so the client never re-derives who defends a
// battle.
//
// The fallback is for frames that predate the field (replays, bug
// reports): a player attack's target IS its defender, and nothing else
// can be read off the id without the rule. `undefined` means nobody
// may block — not attacking, or attacking a planeswalker or battle
// that has left the battlefield (CR 506.4c).
export function defendingPlayerOf(attacker: CardView): string | undefined {
  if (!attacker.attacking_target) return undefined;
  if (attacker.defending_player) return attacker.defending_player;
  const kind = attacker.attacking_target_kind;
  return kind === undefined || kind === "player" ? attacker.attacking_target : undefined;
}

// attackersDefendedBy lists the attackers on the battlefield `seatID`
// defends against: the ones its creatures may be declared as blockers
// for. The block pickers read this, so none of them offers a block the
// server refuses with `not_defending`.
export function attackersDefendedBy(view: GameView, seatID: string): CardView[] {
  return (view.battlefield?.cards ?? []).filter((c) => defendingPlayerOf(c) === seatID);
}
