package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ody_b2_helpers.go — the shapes Odyssey batch 2 (#5613) prints more than
// once: the Eggs, the sacrifice lands, the threshold lands, and the
// Cephalid "target player draws, then discards" rows. Its own file, as
// #231 asks of a batch, so a concurrent batch never collides here.
//
// Append-only: add a builder, never change what one means.

// odyEggRow is the Eggs' one ability, "{2}, {T}, Sacrifice this
// artifact: Add <mana>. Draw a card. (Activate only as an instant.)"
// (Darkwater, Mossfire, Shadowblood, Skycloud and Sungrass Egg). It is
// Chromatic Sphere's shape with two fixed colours: drawing moves a card
// from a library, so it is not a mana ability (CR 605.1a) and uses the
// stack. The costs are paid at announce (CR 602.2b), and on resolution
// the controller adds the mana and then draws, in the printed order.
// "Activate only as an instant" is every activated ability's default
// timing.
func odyEggRow(mana string) ActivatedAbility {
	return ActivatedAbility{
		Label:   "{2}, {T}, Sacrifice this artifact: Add " + mana + ". Draw a card.",
		Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
		Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
		Effect:  Do(AddMana{Produced: mana}, DrawCards{N: 1}),
	}
}

// odySacrificeLand is the whole Spec of Odyssey's five sacrifice lands
// (Abandoned Outpost, Bog Wreckage, Ravaged Highlands, Seafloor Debris,
// Timberland Ruins):
//
//	"This land enters tapped.
//	 {T}: Add <color>.
//	 {T}, Sacrifice this land: Add one mana of any color."
//
// The entry is the plain enters-tapped replacement (CR 614.1d). Both
// rows are mana abilities (CR 605.1a): the second sacrifices the land as
// part of its cost, like Lotus Petal, and the colour is picked as the
// mana is made.
func odySacrificeLand(oracle, name, color string) Spec {
	return Spec{
		OracleID:     oracle,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: color, Label: "Add " + color},
			{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Sacrifice this land: Add one mana of any color",
			},
		},
	}
}

// odyThresholdLand is the whole Spec of an Odyssey threshold land
// (Cabal Pit, Centaur Garden, Nomad Stadium; Barbarian Ring and
// Cephalid Coliseum are the same cycle, written earlier):
//
//	"{T}: Add <color>. This land deals 1 damage to you.
//	 Threshold — <color>, {T}, Sacrifice this land: <clause> Activate only
//	 if there are seven or more cards in your graveyard."
//
// The mana row is a painland half: the damage is a rider, not a cost.
// Threshold is the second row's activation condition (CR 602.1b),
// seven or more cards of any kind. `row` carries the clause's targets,
// purpose and effect; its label and cost are filled in here.
func odyThresholdLand(oracle, name, color, clause string, row ActivatedAbility) Spec {
	row.Label = "Threshold — " + color + ", {T}, Sacrifice this land: " + clause +
		" Activate only if there are seven or more cards in your graveyard."
	row.Cost = Plus(ManaCost(color), TapCost(), SacrificeThis())
	row.Condition = GraveyardAtLeast(7, nil)
	return Spec{
		OracleID:     oracle,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  color,
			Label:     "Add " + color + ". This land deals 1 damage to you.",
			PainToYou: 1,
		}},
		Activated: []ActivatedAbility{row},
	}
}

// odyTargetPlayerLoots is "{T}: Target player draws <n> cards, then
// discards <n> cards." (Cephalid Looter, Cephalid Broker). The target,
// if still legal (CR 608.2b), draws first and then chooses the discards
// from the hand those draws made (drawThenDiscard).
func odyTargetPlayerLoots(label string, n int) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    TapCost(),
		Targets: TargetPlayer("target player"),
		Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: n, Discards: n}),
		Effect: func(g *game.Game, item *game.StackItem) error {
			for _, t := range NewContext(g, item).LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if err := drawThenDiscard(g, t.ID, item.SourceCardID, n, n, ""); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// odyProtectionFromEachColor is "protection from each color": five
// separate protection abilities, one per colour (CR 702.16h).
func odyProtectionFromEachColor() []string {
	out := make([]string, 0, len(game.AllColors))
	for _, c := range game.AllColors {
		out = append(out, game.ProtectionFromColor(c))
	}
	return out
}
