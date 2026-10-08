// A player's speed (CR 702.179, ADR 0138): PlayerView.speed, 1 to 4,
// absent while they have none. The seat chip's hover says the rule in
// plain words, so a player who has never seen Aetherdrift can read it.

/** The speed at which a player has max speed (CR 702.179e). */
export const MAX_SPEED = 4;

/** The hover text for a seat's speed chip. Empty for no speed. */
export function speedTitle(speed: number): string {
  if (speed <= 0) return "";
  if (speed >= MAX_SPEED) {
    return "Max speed (4 of 4). Max speed abilities are on.";
  }
  return `Speed ${speed} of ${MAX_SPEED}. It rises once on each of your turns when an opponent loses life.`;
}
