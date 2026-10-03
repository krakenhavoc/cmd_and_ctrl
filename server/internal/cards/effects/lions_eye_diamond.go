package effects

// Lion's Eye Diamond — Artifact {0}:
//
//	"Discard your hand, Sacrifice this artifact: Add three mana of any
//	 one color. Activate only as an instant."
//
// The card #1600's discard-your-hand cost was built for (ADR 0020's
// 2026-10-02 amendment). Three things make it, and each is one
// declaration:
//
//   - "Discard your hand" is DiscardYourHand's clause on the mana
//     ability's cost: every card in hand goes, nothing is named, and
//     an empty hand pays it (CR 118.3). The cards leave through the
//     one discard door with cause cost, so a madness card is exiled
//     and offered, and a "whenever a player discards" payoff triggers
//     — after the mana is in the pool, because a mana ability resolves
//     at once and its triggers wait for the next time a player would
//     receive priority (CR 605.3b, CR 603.3).
//   - "Add three mana of any one color" is ONE colour pick that adds
//     three of that colour (OneColorOfAmount, Gilded Lotus's grammar).
//   - "Activate only as an instant" is the OnlyAsAnInstant Condition.
//     It is still a mana ability — no stack, no response (CR 605.3b,
//     the 2004-10-04 ruling) — but it may be activated only while the
//     controller could cast an instant: holding priority, owing no
//     prompt, no prompt stopping the table. So it cannot be cracked in
//     the middle of a resolution (to pay a Mana Leak's tax), and the
//     auto-tapper, which runs in the middle of a cast, never plans it:
//     the discard component alone keeps it out of every plan. The play
//     the card is known for — crack it in response with the spell you
//     want already cast, or float the mana and then cast — is open,
//     because the mana stays in the pool until the step ends
//     (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee6099b0-fb1f-42f1-b862-7708c6e36d05",
		Name:         "Lion's Eye Diamond",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Sacrifice:    true,
				DiscardCards: DiscardYourHand().DiscardCards,
			},
			Produced:  OneColorOfAmount(3),
			Label:     "Discard your hand, Sacrifice this artifact: Add three mana of any one color. Activate only as an instant.",
			Condition: OnlyAsAnInstant(),
		}},
	})
}
