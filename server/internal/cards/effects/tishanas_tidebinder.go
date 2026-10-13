package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tishana's Tidebinder — Creature — Merfolk Wizard {2}{U}, 3/2:
//
//	"Flash
//	 When this creature enters, counter up to one target activated or
//	 triggered ability. If an ability of an artifact, creature, or
//	 planeswalker is countered this way, that permanent loses all
//	 abilities for as long as this creature remains on the battlefield.
//	 (Mana abilities can't be targeted.)"
//
// Stifle's clause (AbilityOnStack) with a floor of zero for "up to one".
// A mana ability never uses the stack (CR 605.3b), so the parenthetical
// needs no code. The ability's source is read before the counter moves
// anything: if it is an artifact, a creature or a planeswalker on the
// battlefield, it loses all abilities (a scoped layer-6 effect) for as
// long as this creature remains on the battlefield (CR 611.2b). A
// Tidebinder that has left or been replaced by the time its trigger
// resolves starts no effect, and one that leaves later hands the
// abilities back. An ability whose source is not such a permanent (a
// land's ability, a spell's cast trigger, a source that has left the
// battlefield) is only countered, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2993dc7d-723d-4a9b-94bd-4bb02a9f7243",
		Name:            "Tishana's Tidebinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, Self,
					"Tishana's Tidebinder — counter up to one target activated or triggered ability",
					tidebinderETB),
				AbilityOnStack("up to one target activated or triggered ability").WithCount(0, 1)),
		},
	})
}

func tidebinderETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		victim := g.StackItemForEffect(t.ID)
		if victim == nil {
			continue
		}
		source := victim.SourceCardID
		if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
			return err
		}
		c, ok := g.LookupCardForEffect(source)
		if !ok || !onBattlefield(g, source) {
			continue
		}
		if !c.IsArtifact() && !c.IsCreature() && !c.IsPlaneswalker() {
			continue
		}
		d, ok := DurationWhileSourceRemains(ctx, item.SourceCardID)
		if !ok {
			continue
		}
		if err := (ScopedEffectFor{
			Target:   source,
			Mods:     []game.Mod{game.LoseAllAbilitiesMod()},
			Duration: d,
			Label:    "Tishana's Tidebinder — loses all abilities",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
