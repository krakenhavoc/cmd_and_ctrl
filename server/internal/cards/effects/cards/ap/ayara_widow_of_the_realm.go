package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ayara, Widow of the Realm // Ayara, Furnace Queen — {1}{B}{B} Legendary
// Creature — Elf Noble 3/3 // Legendary Creature — Phyrexian Elf Noble 4/4
// (March of the Machine):
//
//	Front: "{T}, Sacrifice another creature or artifact: Ayara deals X
//	        damage to target opponent or battle and you gain X life, where
//	        X is the sacrificed permanent's mana value.
//	        {5}{R/P}: Transform Ayara. Activate only as a sorcery."
//	Back:  "At the beginning of combat on your turn, return up to one
//	        target artifact or creature card from your graveyard to the
//	        battlefield. It gains haste. Exile it at the beginning of the
//	        next end step."
//
// X is the sacrificed permanent's mana value as it last existed on the
// battlefield (CR 608.2h). The life is gained whether or not the damage
// was prevented. The back's return is "up to one", so it needs no target;
// the returned card gains haste and is exiled at the next end step.
//
// No simplification.
func init() {
	const oracle = "f9ef90b9-a2b4-4a5b-92fb-2268ea7d709f"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Ayara, Widow of the Realm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice another creature or artifact: Ayara deals X damage to target opponent or battle and you gain X life, where X is the sacrificed permanent's mana value.",
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeAnotherN(1, "another creature or artifact", effects.Or(effects.Creature(), effects.Artifact()))),
				Targets: effects.S82Mom16TargetOpponentOrBattle(),
				Effect:  ayaraWidowDrain,
			},
			{
				Label:        "{5}{R/P}: Transform Ayara. Activate only as a sorcery.",
				Cost:         effects.ManaCost("{5}{R/P}"),
				SorcerySpeed: true,
				Effect:       effects.Do(effects.TransformThis{}),
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Ayara, Furnace Queen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Ayara, Furnace Queen — return up to one target artifact or creature card from your graveyard to the battlefield; it gains haste; exile it at the beginning of the next end step",
				ayaraFurnaceQueenReturn),
				effects.TargetCardInGraveyard("up to one target artifact or creature card from your graveyard", effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 1)),
		},
	})
}

func ayaraWidowDrain(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := effects.SacrificedManaValue(ctx)
	if len(ctx.LegalTargets()) == 0 {
		return nil
	}
	if err := effects.SourceDealsComputedDamageToEachLegalTarget(func(*effects.Context) int { return x })(g, item); err != nil {
		return err
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
}

func ayaraFurnaceQueenReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false, nil, ctx.ItemTail(ayaraFurnaceQueenEntered))
}

// ayaraFurnaceQueenEntered gives the returned card haste and exiles it at
// the next end step. An on-disk key: never renamed, never reused.
var ayaraFurnaceQueenEntered = effects.EnteredTail("ayara-furnace-queen/entered", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !effects.OnBattlefield(g, entered) {
		return nil
	}
	if err := (effects.ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Ayara, Furnace Queen — it gains haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		Label: "Ayara, Furnace Queen — exile it",
		Cards: []uuid.UUID{entered},
		Body:  effects.ExileListedCardsBody,
	}.Apply(ctx)
})
