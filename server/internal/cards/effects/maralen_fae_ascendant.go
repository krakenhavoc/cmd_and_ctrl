package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maralen, Fae Ascendant — Legendary Creature — Elf Faerie Noble
// {2}{B}{G}{U}, 4/5 (Lorwyn Eclipsed, #4458):
//
//	"Flying
//	 Whenever Maralen or another Elf or Faerie you control enters, exile
//	 the top two cards of target opponent's library.
//	 Once each turn, you may cast a spell with mana value less than or
//	 equal to the number of Elves and Faeries you control from among
//	 cards exiled with Maralen this turn without paying its mana cost."
//
// The trigger is High Perfect Morcant's "this or another <type> you
// control enters" over Elf or Faerie (words of the card's text, CR
// 612.2), so Maralen entering with other Elves and Faeries triggers once
// for each of them (the ruling). It exiles the top two cards of the
// targeted opponent's library face up, each LINKED to this Maralen
// object (CR 607.2a, Game.ExileLinkedForEffect), which also stamps the
// turn of the link (Card.ExiledWithTurn).
//
// The cast is a gated STANDING permission over exile (Azula, Cunning
// Usurper's shape), derived off the battlefield on every query, so it
// ends the moment Maralen leaves and a Maralen that returns is a new
// object with no claim on its earlier self's cards (CR 400.7):
//
//   - "Cards exiled with Maralen this turn" is ExiledWithSource narrowed
//     by ExiledWithSourceThisTurn (the #4458 seam): a card linked on an
//     earlier turn is not offered.
//   - "Mana value less than or equal to the number of Elves and Faeries
//     you control" is MaxSpellManaValueCountOf (the same seam): the
//     derivation counts the permanents you control that are an Elf or a
//     Faerie, each once, as the spell is cast, and the cap is judged on
//     the face being cast with X at 0 (the ruling).
//   - "Without paying its mana cost" is the permission's flat {0} price
//     (Court of Locthwain's), so X is 0 (CR 107.3b, the ruling), no other
//     alternative cost applies, and additional costs are still paid
//     (CR 118.9a, the rulings). It is cast only, never a land play.
//   - "Once each turn" is per Maralen object: the cast follow-up records
//     the use (Game.NoteDoneThisTurn, Emet-Selch's shape) once the spell
//     is on the stack through this permission, and the gate withholds the
//     permission for the rest of the turn.
//
// No simplification.
const (
	maralenFaeAscendantName = "Maralen, Fae Ascendant"
	// maralenFaeAscendantDoneLabel is the tally the once-each-turn limit
	// reads.
	maralenFaeAscendantDoneLabel = "Maralen, Fae Ascendant — free cast (done)"
	// maralenFaeAscendantFreeCastKey is the permission's {0} price, as
	// a cast names it (CastSpellParams.AlternativeCost).
	maralenFaeAscendantFreeCastKey = "maralen_fae_ascendant"
)

// maralenFaeAscendantCastTally spends the turn's use once a spell is cast
// through the permission. An on-disk key.
var maralenFaeAscendantCastTally = game.RegisterCastFollowUp("maralen-fae-ascendant-once-each-turn",
	func(g *game.Game, f game.CastFollowUp) error {
		g.NoteDoneThisTurn(f.Source, maralenFaeAscendantDoneLabel)
		return nil
	})

func init() {
	Register(Spec{
		OracleID:        "114e1d2c-cefc-4e2a-922e-3b2ceed1871d",
		Name:            maralenFaeAscendantName,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, maralenOrAnotherElfOrFaerieEntered,
					"Maralen, Fae Ascendant — exile the top two cards of target opponent's library",
					maralenFaeAscendantExileTopTwo),
				TargetPlayer("target opponent", Opponent()),
			),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				return g.ResolvedThisTurn(source, maralenFaeAscendantDoneLabel) == 0
			},
			Permission: game.CastPermission{
				Zone: game.ZoneExile,
				Filter: game.PermissionFilter{
					ExiledWithSource:         true,
					ExiledWithSourceThisTurn: true,
				},
				CastOnly:                 true,
				MaxSpellManaValueCountOf: []string{"Elf", "Faerie"},
				AltCostKey:               maralenFaeAscendantFreeCastKey,
				Cost:                     "{0}",
				FollowUp:                 maralenFaeAscendantCastTally,
				Label:                    "Once each turn, cast a spell with mana value up to the number of Elves and Faeries you control from among cards exiled with Maralen this turn without paying its mana cost (Maralen, Fae Ascendant)",
			},
		}},
	})
}

// maralenOrAnotherElfOrFaerieEntered is "whenever Maralen or another Elf
// or Faerie you control enters".
func maralenOrAnotherElfOrFaerieEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok {
		return false
	}
	return c.InstanceID == source.InstanceID ||
		c.HasSubtype(source.TextSubtype("Elf")) || c.HasSubtype(source.TextSubtype("Faerie"))
}

// maralenFaeAscendantExileTopTwo exiles the top two cards of the still
// legal target opponent's library, linked to this Maralen object. A
// library with fewer cards exiles what it has.
func maralenFaeAscendantExileTopTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := TargetedPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	// A Maralen that has already left still exiles the cards; they are
	// linked to no object then, and nothing can cast them through her.
	src, _ := g.SourceObjectForEffect(item)
	p := g.PlayerByIDForEffect(opp)
	if p == nil || p.Library == nil {
		return nil
	}
	n := len(p.Library.Cards)
	var top []uuid.UUID
	for i := n - 1; i >= 0 && i >= n-2; i-- {
		top = append(top, p.Library.Cards[i].InstanceID)
	}
	for _, id := range top {
		if err := g.ExileLinkedForEffect(src, id); err != nil {
			return err
		}
	}
	return nil
}
