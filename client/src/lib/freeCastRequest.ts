import { guardedWritable } from "./guardedStore";
import type { CardView } from "./protocol";

// ADR 0099 §7. Answering "Cast it free" to a discover's or a cascade's
// may_cast prompt only opens the free cast: the server stamps a grant on
// the exiled card, and the cast itself is an ordinary cast_spell. The
// prompt lives in ChoicePromptModal and the one cast entry point
// (handlePlayCard, #874) lives in Board, so the modal leaves the card's
// instance id here and Board starts the cast chain — face picker,
// targets, modes — once a snapshot shows the card in exile and castable.
//
// The window closes on the player's next pass, so a request that never
// becomes castable (no legal target, say) is dropped rather than kept:
// the player passes, and the card goes where the keyword sends it.
export const freeCastRequest = guardedWritable<string | null>(null, "freeCastRequest");

// MayCastKeywordsThatOpenACast are the may_cast keywords whose accept
// branch is a free cast from exile, as opposed to madness's paid cast
// or an offer the card's own text words.
export const mayCastKeywordsThatOpenACast: ReadonlySet<string> = new Set([
  "discover",
  "cascade",
  "suspend",
]);

// freeCastTarget is Board's half: the exiled card a pending request
// names, when it is there and the viewer may cast it now. `undefined`
// means "not yet" (the snapshot carrying the grant has not arrived);
// `null` means the request is stale and should be dropped (the card is
// in exile but nobody may cast it).
export function freeCastTarget(
  exile: CardView[] | undefined,
  id: string | null,
): CardView | null | undefined {
  if (!id) return null;
  const card = exile?.find((c) => c.instance_id === id);
  if (!card) return undefined;
  return card.castable_here ? card : null;
}
