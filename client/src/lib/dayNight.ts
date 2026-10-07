// dayNight.ts — how the dock says the game is day or night (CR 731,
// ADR 0132). The designation belongs to the game, not to a seat, so
// there is one chip and it reads the same for every viewer.

/** The wire value of GameView.day_night: absent while the game has neither. */
export type DayNight = "day" | "night";

export interface DayNightChip {
  /** The word on the chip. */
  label: string;
  /** A glyph, aria-hidden: the word carries the meaning. */
  glyph: string;
  /** The hover and screen-reader explanation, in the rule's own terms. */
  title: string;
}

/**
 * The chip for a designation, or null while the game is neither day nor
 * night (nothing to say, and a "neither" chip would be noise at the many
 * tables that never use a day/night card). Anything the server might one
 * day send that this build does not know is null too, rather than a
 * chip that claims the wrong thing.
 */
export function dayNightChip(value: string | undefined): DayNightChip | null {
  switch (value) {
    case "day":
      return {
        label: "Day",
        glyph: "☀",
        title:
          "It is day. It becomes night at the start of a turn if the previous turn's player cast no spells.",
      };
    case "night":
      return {
        label: "Night",
        glyph: "☾",
        title:
          "It is night. It becomes day at the start of a turn if the previous turn's player cast two or more spells.",
      };
    default:
      return null;
  }
}
