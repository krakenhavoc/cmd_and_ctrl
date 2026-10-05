// ADR 0107 §6 (#1860): the caption under each candidate of a
// "choose_source" prompt — "a source of your choice" (CR 609.7a).
//
// The candidates come from several zones at once: permanents, spells on
// the stack, face-up cards in a command zone, and cards in a graveyard or
// exile that something on the stack (or a waiting effect) still refers
// to. The card face alone does not say which, nor whose it is, and both
// decide the choice: a Lightning Bolt on the stack is about to deal its
// damage, a creature in a graveyard is the source of the trigger that
// will. So each card gets "<controller>'s · <where>".

import type { CardView, GameView } from "./protocol";

function inZone(cards: CardView[] | undefined, id: string): boolean {
  return (cards ?? []).some((c) => c.instance_id === id);
}

// damageSourceWhere names the zone the card is in, as a phrase.
export function damageSourceWhere(snap: GameView, id: string): string {
  if (inZone(snap.battlefield?.cards, id)) return "on the battlefield";
  if (inZone(snap.stack?.cards, id)) return "on the stack";
  if (inZone(snap.exile?.cards, id)) return "in exile";
  for (const seat of snap.seats ?? []) {
    if (inZone(seat.command?.cards, id)) return "in the command zone";
    if (inZone(seat.graveyard?.cards, id)) return "in a graveyard";
  }
  return "";
}

// damageSourceCaption is the full caption: whose, and where. The viewer's
// own sources say "Yours", so a seat can see at a glance which candidates
// are threats.
export function damageSourceCaption(
  snap: GameView,
  card: CardView,
  viewerID: string | null,
): string {
  let whose = "";
  if (card.controller) {
    if (viewerID && card.controller === viewerID) {
      whose = "Yours";
    } else {
      const seat = (snap.seats ?? []).find((s) => s.id === card.controller);
      whose = seat ? `${seat.name}'s` : "";
    }
  }
  const where = damageSourceWhere(snap, card.instance_id);
  return [whose, where].filter(Boolean).join(" · ");
}

// permanentWhoseCaption is the short "whose is this?" line under a
// battlefield candidate in a card-set pick that spans several seats —
// Peregrine Drake's "untap up to five lands" lists every tapped land at
// the table, and a grid of identical Forests does not say which are the
// viewer's (#1960). Empty for a card that is not on the battlefield, so
// a hand or graveyard pick keeps its plain face.
export function permanentWhoseCaption(
  snap: GameView,
  card: CardView,
  viewerID: string | null,
): string {
  if (!inZone(snap.battlefield?.cards, card.instance_id) || !card.controller) return "";
  if (viewerID && card.controller === viewerID) return "Yours";
  const seat = (snap.seats ?? []).find((s) => s.id === card.controller);
  return seat ? `${seat.name}'s` : "";
}
