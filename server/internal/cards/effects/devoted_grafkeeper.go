package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devoted Grafkeeper // Departed Soulkeeper (#1855, ADR 0107 §4) — a
// disturb card.
//
// Front face, Creature — Human Peasant {W}{U}, 2/1:
//
//	"When this creature enters, mill two cards.
//	 Whenever you cast a spell from your graveyard, tap target creature
//	 you don't control.
//	 Disturb {1}{W}{U} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit, 3/1:
//
//	"Flying
//	 This creature can block only creatures with flying.
//	 If Departed Soulkeeper would be put into a graveyard from
//	 anywhere, exile it instead."
//
// "A spell from your graveyard" is read off the spell's announce
// record and its owner (youCastASpellFromYourGraveyard), so a
// flashback, an escape and another disturb all count, and a spell cast
// out of an opponent's graveyard does not. Disturbing the Grafkeeper
// itself does not trigger it: it is not on the battlefield then.
//
// "Can block only creatures with flying" is a block rule on the
// blocker's side of the pair (CantBlockAttackers), refusing every
// attacker without flying, read live as blockers are declared.
//
// No simplification.
func init() {
	tap := On(game.EventCast, youCastASpellFromYourGraveyard,
		"Devoted Grafkeeper — tap target creature you don't control", b36TapChosenCreature)
	tap.Targets = TargetCreature("target creature you don't control", OpponentControls())
	Register(Spec{
		OracleID:         devotedGrafkeeperOracleID,
		Name:             "Devoted Grafkeeper",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{1}{W}{U}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Devoted Grafkeeper — mill two cards", Do(MillCards{N: 2})),
			tap,
		},
	})
	Register(Spec{
		OracleID:        devotedGrafkeeperOracleID + "#1",
		Name:            "Departed Soulkeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Departed Soulkeeper")},
	})
}

const devotedGrafkeeperOracleID = "0a154fb2-9f23-4c22-baee-728492385d6d"
