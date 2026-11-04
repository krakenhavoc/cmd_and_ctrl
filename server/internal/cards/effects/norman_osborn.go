package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Norman Osborn // Green Goblin — a modal double-faced card whose front
// face transforms (Marvel's Spider-Man promos, #4460).
//
//	Norman Osborn                       {1}{U}   1/1
//	Legendary Creature — Human Scientist Villain
//	  Norman Osborn can't be blocked.
//	  Whenever Norman Osborn deals combat damage to a player, he connives.
//	  {1}{U}{B}{R}: Transform Norman Osborn. Activate only as a sorcery.
//
//	Green Goblin                        {1}{U}{B}{R}   3/3
//	Legendary Creature — Goblin Human Villain
//	  Flying, menace
//	  Spells you cast from your graveyard cost {2} less to cast.
//	  Goblin Formula — Each nonland card in your graveyard has mayhem. The
//	  mayhem cost is equal to its mana cost. (You may cast a card from your
//	  graveyard for its mayhem cost if you discarded it this turn. Timing
//	  rules still apply.)
//
// The front face is Gwen Stacy's shape: the S24 "can't be blocked" bit, the
// connive (CR 701.50) of the creature that dealt the damage, and
// TransformThis at sorcery speed (CR 712.9 admits a modal double-faced
// permanent an instruction transforms). The back face registers under the
// oracle ID plus "#1" (game.CatalogKey).
//
// Green Goblin's reduction is Emet-Selch's: it prices every cast its
// controller makes from a graveyard, whatever let the card be cast there.
// Goblin Formula is Underworld Breach's standing permission (ADR 0066):
// derived from the battlefield on every query, so a card has mayhem while
// it is in the graveyard and Green Goblin is on the battlefield, and two
// Goblins compose. The offer is priced at the card's own mana cost and is
// claimable only for a card discarded this turn (CR 702.187b,
// game.AltCostKeyGrantedMayhem); a card that prints mayhem keeps its own
// offer beside it and the caster picks (CR 113.2c). Timing rules still
// apply: the permission names no timing of its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     normanOsbornOracle,
		Name:         "Norman Osborn",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Norman Osborn — he connives", Do(Instruction(sourceConnives))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{U}{B}{R}: Transform Norman Osborn. Activate only as a sorcery.",
			Cost:         ManaCost("{1}{U}{B}{R}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        normanOsbornOracle + "#1",
		Name:            "Green Goblin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Spells you cast from your graveyard cost {2} less to cast.",
				YourSpell(), CastFromGraveyard()),
		},
		CastPermissions: []game.CastPermission{{
			Zone: game.ZoneGraveyard,
			// "Each nonland CARD in your graveyard", as Underworld Breach
			// reads it (#4509).
			Filter:     game.PermissionFilter{CardInZone: true, NonLandOnly: true},
			AltCostKey: game.AltCostKeyGrantedMayhem,
			Label:      "Goblin Formula — mayhem: its mana cost, if you discarded it this turn",
		}},
	})
}

// normanOsbornOracle is shared by the two faces.
const normanOsbornOracle = "aa5b06a4-90c3-45e7-ab83-cb5c96ded21d"
