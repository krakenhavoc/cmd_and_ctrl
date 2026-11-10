package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawn-Blessed Pennant — Artifact {1} (Lorwyn Eclipsed, #4458):
//
//	"As this artifact enters, choose Elemental, Elf, Faerie, Giant,
//	 Goblin, Kithkin, Merfolk, or Treefolk.
//	 Whenever a permanent you control of the chosen type enters, you gain
//	 1 life.
//	 {2}, {T}, Sacrifice this artifact: Return target card of the chosen
//	 type from your graveyard to your hand."
//
// The choice is the CR 614.12 creature-type choice limited to the eight
// printed types (ChooseCreatureTypeFromAsEnters), landing on
// Card.NamedTribe as every chosen type does. The life trigger is Kindred
// Discovery's shape for any permanent, read through the layered view (a
// changeling or a kindred permanent of the type counts), with the
// footprinted GainLife primitive. Before the type is chosen nothing
// matches.
//
// The return is Codex Shredder's sacrifice-for-a-card, narrowed to the
// chosen type with OfTheChosenType: the Pennant is sacrificed as the
// cost, after the target is picked (CR 601.2c before 601.2h), and the
// target is re-checked on resolution against the type the Pennant had as
// it last existed (CR 608.2h). A changeling card is every type in the
// graveyard too (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "311eef28-21e9-44fa-96df-d281f0a9ba34",
		Name:         "Dawn-Blessed Pennant",
		Completeness: CompletenessFull,
		AsEnters: ChooseCreatureTypeFromAsEnters("Dawn-Blessed Pennant",
			"Elemental", "Elf", "Faerie", "Giant", "Goblin", "Kithkin", "Merfolk", "Treefolk"),
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, dawnBlessedPennantOfChosenTypeEntered,
				"Dawn-Blessed Pennant — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this artifact: Return target card of the chosen type from your graveyard to your hand.",
			Cost:  Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: OfTheChosenType(TargetCardInGraveyard(
				"target card of the chosen type from your graveyard", YouOwn())),
			Effect: returnFirstLegalGraveyardTargetToHand,
		}},
	})
}

// dawnBlessedPennantOfChosenTypeEntered is "a permanent you control of
// the chosen type enters".
func dawnBlessedPennantOfChosenTypeEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source.NamedTribe == "" {
		return false
	}
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtype(source.NamedTribe)
}
