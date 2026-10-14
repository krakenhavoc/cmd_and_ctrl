package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Null Summoner — Creature — Human Warlock {2}{U}{B}, 4/2 (Reality
// Fracture):
//
//	"When this creature enters, if you cast it, target opponent reveals
//	 their hand. You choose a nonland card from it. Exile that card.
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, you may cast the exiled card, and mana of any type can
//	 be spent to cast that spell."
//
// Three engine pieces meet here (#3516).
//
// The LINK. The enters trigger is a revealed-hand pick that exiles
// (ADR 0116), with `Linked: true`: the pick takes the resolving item's
// source OBJECT when it goes up and, once the chosen card has landed in
// exile, stamps it as Card.ExiledWith (CR 607.2a, ADR 0155 §8), through
// the same exile route a linked exile takes. So "the exiled card" is
// this Summoner's card and no other: a card another effect exiled from
// the same hand, or one this Summoner's earlier incarnation exiled, is
// not linked to this object (CR 400.7).
//
// The PERMISSION. A standing cast permission over exile, derived off the
// battlefield on every query and never stored (Valgavoth, Terror Eater's
// shape, ADR 0066), narrowed by `ExiledWithSource`: the card's link must
// name THIS Summoner. It reaches the card however long ago it was exiled,
// and ends the moment the Summoner leaves, which is the printed "as long
// as": a Summoner that has left leaves the card stranded in exile.
//
// The GATE. Threshold is an ability word with no rules of its own
// (CR 207.2c); the clause is a graveyard count, so the permission is a
// GATED standing one (`Spec.GatedCastPermissions`, `GraveyardAtLeast(7)`,
// the gate Fortune Teller's Talent's second line uses). It is derived on
// every query, so while fewer than seven cards are in your graveyard it
// does not exist: the card stays exiled with its link, and the cast opens
// the moment the count reaches seven and closes if it falls back. The
// count is read live, when you try to cast.
//
// "Mana of any type can be spent" is `AnyType`, the wider clause of the
// two (colorless is a type of mana but not a colour, CR 106.1b), applied
// to the spell cast through this permission only. "May cast" leaves
// timing alone: an exiled instant is cast at instant speed, a creature at
// sorcery speed, and its mana cost is paid (no alternative cost). The
// permission is CastOnly: the card was chosen as a nonland card, and a
// modal double-faced card's land face is played, not cast.
//
// "If you cast it" is an intervening if (CR 603.4) read off the
// permanent's cast provenance (CR 400.7d), as Geological Appraiser's: a
// Summoner returned from a graveyard or flickered in was not cast as it
// entered and exiles nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56c412be-daab-4461-931f-1e2ea8ecd75d",
		Name:         "Null Summoner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, nullSummonerWasCast),
					"Null Summoner — target opponent reveals their hand, you choose a nonland card and exile it",
					nullSummonerExileFromHand),
				TargetPlayer("target opponent", Opponent())),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: GraveyardAtLeast(7, nil),
			Permission: game.CastPermission{
				Zone:     game.ZoneExile,
				Filter:   game.PermissionFilter{ExiledWithSource: true},
				AnyType:  true,
				CastOnly: true,
				Label:    "Threshold — cast the card exiled with Null Summoner; mana of any type can be spent (Null Summoner)",
			},
		}},
	})
}

// nullSummonerWasCast is "if you cast it": the entering permanent came
// from a spell.
func nullSummonerWasCast(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.FromZone != ""
}

// nullSummonerExileFromHand is the trigger's effect: the targeted
// opponent reveals their hand, the controller chooses a nonland card
// from it, and that card is exiled linked to this Summoner.
func nullSummonerExileFromHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ChooseFromRevealedHand{
		Player: TargetedPlayer(ctx),
		Filter: Nonland(),
		Label:  "nonland card",
		Exile:  true,
		Linked: true,
	}.Apply(ctx)
}
