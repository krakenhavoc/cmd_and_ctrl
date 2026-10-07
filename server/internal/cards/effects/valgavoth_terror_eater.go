package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valgavoth, Terror Eater — Legendary Creature — Elder Demon
// {6}{B}{B}{B}, 9/9:
//
//	"Flying, lifelink
//	 Ward—Sacrifice three nonland permanents.
//	 If a card you didn't control would be put into an opponent's
//	 graveyard from anywhere, exile it instead.
//	 During your turn, you may play cards exiled with Valgavoth. If
//	 you cast a spell this way, pay life equal to its mana value
//	 rather than pay its mana cost."
//
// The nine-drop that eats the table's graveyards. All four clauses
// ship.
//
// The WARD is the reason WardSacrificeCost grew a count: every ward
// in the catalog before this one demanded a single permanent, and the
// pick was hardcoded to one. It is one payment (CR 118.3), so a payer
// with only two nonland permanents cannot pay at all — the spell is
// countered with no prompt, rather than taking two and letting the
// removal through.
//
// The REPLACEMENT is the shared GraveyardBecomesExile with both of
// its narrowing clauses set. "Into an OPPONENT's graveyard" is
// OpponentsOnly; "a card YOU DIDN'T CONTROL" is
// NotControlledByYou, and it is not redundant — a creature you stole
// from an opponent goes to its owner's graveyard when it dies, which
// OpponentsOnly alone would exile and printed Valgavoth would not.
// Being stronger than printed is the one direction a simplification
// may never go, so the second clause is a field rather than a comment.
//
// It catches every graveyard arrival since #931 — death, mill,
// discard, a library search that bins the card, surveil — because
// they all push their move through the CR 614 window. The
// battlefield case is the one worth saying out loud: an opponent's
// creature that would DIE is exiled instead, so it never died and its
// own dies-triggers never fire (CR 700.4).
//
// The FOURTH clause (#2530) was held back for a while, and the answer
// is that it needs no per-card permission at all. A replacement runs
// before the card moves, so it cannot register a permission over the
// card it is moving — the #1117 triage. Two halves instead, each
// already shaped like something the engine had:
//
//   - The LINK. LinkExiled makes the replacement leave
//     ReplacementEvent.ExiledWith on the event, and the move that
//     lands the card in exile stamps it as Card.ExiledWith: Valgavoth
//     as the object it is now (CR 607.2a, 400.7). A card a different
//     replacement exiled (Leyline of the Void) or an ability exiled is
//     not linked, because only this replacement declares it.
//   - The PERMISSION. A STANDING cast permission over exile, derived
//     off the battlefield on every query and never stored (Tinybones,
//     Bauble Burglar's shape, ADR 0066), narrowed by the filter
//     ExiledWithSource: the card's link must name THIS Valgavoth. So
//     it reaches cards exiled before anyone asked, ends the moment
//     Valgavoth leaves, and a Valgavoth that comes back is a new
//     object with no claim on the old one's cards.
//
// "During your turn" is TimingYourTurnOnly, which refuses the cast off
// the holder's own turn and leaves the card's own timing in force, so
// an exiled creature is still a sorcery-speed cast. Lands are played,
// not cast, so no cost attaches to them (CastOnly is deliberately
// unset, and the alternative cost is never offered on a land).
//
// "Pay life equal to its mana value rather than pay its mana cost" is
// the alternative cost Bolas's Citadel already claims
// (LifeEqualToManaValue, CR 118.9 and 119.4): the price is life and no
// mana, it is a cost so a player without the life cannot claim it, and
// a spell with mana value 0 costs nothing. It is the only way to cast
// a card this way — the printed mana cost is not on offer, because the
// clause says "rather than".
//
// Whose card it is does not matter: the cards are the opponents', and
// Valgavoth's controller plays them from exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cae3ec72-436d-4086-9dcb-17b3d92ad5c4",
		Name:            "Valgavoth, Terror Eater",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Ward(WardSacrificeN(3, "three nonland permanents", Nonland()),
				"Valgavoth, Terror Eater — ward, sacrifice three nonland permanents"),
		},
		Replacements: []game.ReplacementEffect{GraveyardBecomesExile{
			OpponentsOnly:      true,
			NotControlledByYou: true,
			LinkExiled:         true,
			Label:              "Valgavoth, Terror Eater — exile it instead",
		}.Build()},
		CastPermissions: []game.CastPermission{{
			Zone:                 game.ZoneExile,
			Filter:               game.PermissionFilter{ExiledWithSource: true},
			Timing:               game.TimingYourTurnOnly,
			AltCostKey:           "valgavoth_terror_eater",
			LifeEqualToManaValue: true,
			Label:                "Pay life equal to its mana value (Valgavoth, Terror Eater)",
		}},
	})
}
