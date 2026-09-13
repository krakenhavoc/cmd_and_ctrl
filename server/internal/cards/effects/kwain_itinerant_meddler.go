package effects

// Kwain, Itinerant Meddler — Legendary Creature — Rabbit Wizard
// {W}{U}, 1/3 (EDHREC rank 2151):
//
//	"{T}: Each player may draw a card, then each player who drew a
//	 card this way gains 1 life."
//
// The group-hug Rabbit. A tap ability (summoning sickness applies, CR
// 302.6) whose body walks the table APNAP: each player draws, then
// each player who drew gains 1 — two passes, in the printed order, so
// every "whenever a player draws" watcher fires before any life
// changes.
//
// Sandbox simplification, declared (the Arcane Denial posture): the
// per-player "may" is not offered. The engine's only resolution-time
// yes/no prompts are the trigger prompt (a card's own controller
// only) and the pay / don't-pay dialog, whose copy is about mana and
// which the engine itself declines to reuse as a bare yes/no. So
// every player draws — except a player whose library is empty, who is
// treated as declining (the one case the printed choice is ever
// used the other way) and gains nothing. A player who would have
// declined a full-library draw to dodge an Underworld Dreams cannot;
// declared on the card.
func init() {
	Register(Spec{
		OracleID:     "ce6a85b5-e249-4582-a6be-ce200da5fd53",
		Name:         "Kwain, Itinerant Meddler",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Every player draws and gains the life — nobody is asked whether they want to, though a player with no library left is skipped."},
		Activated: []ActivatedAbility{{
			Label:  "{T}: Each player draws a card, then each player who drew gains 1 life",
			Cost:   TapCost(),
			Effect: b20EachPlayerDrawsAndGainsOne,
		}},
	})
}
