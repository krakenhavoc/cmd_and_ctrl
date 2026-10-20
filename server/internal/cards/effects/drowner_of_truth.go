package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drowner of Truth // Drowned Jungle — Modal DFC. Front: Creature —
// Eldrazi {5}{G/U}{G/U}, 7/6:
//
//	"Devoid
//	 When you cast this spell, if {C} was spent to cast it, create two
//	 0/1 colorless Eldrazi Spawn creature tokens with 'Sacrifice this
//	 token: Add {C}.'"
//
// Back: Land (mdfc_lands.go).
//
// Devoid is the printed keyword (colourless, CR 702.114). The cast
// trigger is an intervening if (CR 603.4): AppliesTo reads the mana the
// spell was paid with off its stack item at the moment it is cast, so a
// spell paid with no colourless mana puts nothing on the stack. A payment
// the engine waived is unknown and counts as no {C} (the weaker answer).
// ReadsOwnPayment lets a waived payment say what it was paid on paper, and
// WantsColors asks the auto-tapper to spend colourless mana when it has a
// free choice (an ordering, never a filter). The trigger resolves above
// the spell, so the Spawn arrive even if the Drowner is countered. No
// simplification.
func init() {
	cast := WhenYouCastThisSpell("Drowner of Truth — create two Eldrazi Spawn tokens",
		Do(CreateToken{Template: EldraziSpawnToken(), N: 2}))
	cast.AppliesTo = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.CardID == source.InstanceID && ev.Actor == source.Controller &&
			g.StackItemPaidForEffect(ev.CardID).Spent().Count("C") > 0
	}
	Register(Spec{
		OracleID:        "db19a27a-ee22-4931-ae3c-0ce21f456ea6",
		Name:            "Drowner of Truth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		ReadsManaSpent:  game.ReadsOwnPayment,
		WantsColors:     []string{"C"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}
