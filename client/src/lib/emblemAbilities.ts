// emblemAbilities.ts — an emblem's activated abilities on the table
// (#5868, CR 114.4). The server sends the rows on the emblem's
// `zone_abilities`, to its owner alone; a click hands Board's announce
// chain a card-shaped stand-in for the emblem, so the activation is the
// ordinary `activate_ability` with the emblem's instance_id as
// `source_card_id`, and every cost and target picker is the one a card uses.
import type { ActivatedAbilityView, CardView, EmblemView } from "./protocol";

/** The emblem as the announce chain reads a source: its id, its label
 *  for a name, its owner as controller (CR 114.2), and its rows. */
export function emblemAsCard(emblem: EmblemView, ownerID: string): CardView {
  return {
    instance_id: emblem.instance_id,
    name: emblem.label,
    owner: ownerID,
    controller: ownerID,
    zone_abilities: emblem.zone_abilities,
  };
}

/** The cost half of a "cost: effect" row, for the button's name. */
export function emblemAbilityCost(row: ActivatedAbilityView): string {
  const label = row.label ?? "";
  const at = label.indexOf(":");
  return at > 0 ? label.slice(0, at).trim() : label || "Activate";
}

/** Why a row can't be activated now, or "" when it can. */
export function emblemAbilityBlocked(row: ActivatedAbilityView, hasPriority: boolean): string {
  if (row.cant_activate) return row.cant_activate;
  if (row.timing_closed) return "not now";
  if (!hasPriority) return "you don't hold priority";
  return "";
}
