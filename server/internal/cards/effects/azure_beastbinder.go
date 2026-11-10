package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azure Beastbinder — Creature — Rat Rogue {1}{U}, 1/3 (Bloomburrow,
// #4467):
//
//	"Vigilance
//	 This creature can't be blocked by creatures with power 2 or greater.
//	 Whenever this creature attacks, up to one target artifact, creature, or
//	 planeswalker an opponent controls loses all abilities until your next
//	 turn. If it's a creature, it also has base power and toughness 2/2
//	 until your next turn."
//
// Vigilance rides PrintedKeywords. The evasion is a block rule on the
// Beastbinder (CantBeBlockedBy over power 2 or greater, read when blocks
// are declared; raising a blocker's power afterwards does not unblock it,
// the ruling). The attack trigger targets up to one opposing permanent; as
// it resolves, one scoped effect (pinned to the object, until the start of
// your next turn) removes all its abilities (layer 6, CR 613.1f), and if it
// is a creature then the same effect sets its base power and toughness to
// 2/2 (layer 7b), under whatever pumps and counters apply afterwards. It
// may later gain abilities (the ruling), since the removal is timestamp
// ordered. A target that is gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7bb51602-60e8-4730-b11c-2870d6b42e20",
		Name:            "Azure Beastbinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerGE(2), "creatures with power 2 or greater"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Azure Beastbinder — up to one target artifact, creature, or planeswalker an opponent controls loses all abilities until your next turn; if it's a creature, base power and toughness 2/2",
				azureBeastbinderBind),
				TargetPermanent("up to one target artifact, creature, or planeswalker an opponent controls",
					Or(Artifact(), Creature(), Planeswalker()), OpponentControls()).WithCount(0, 1)),
		},
	})
}

func azureBeastbinderBind(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	mods := []game.Mod{game.LoseAllAbilitiesMod()}
	if c.IsCreature() {
		mods = append(mods, game.SetBasePTMods(2, 2)...)
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     mods,
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Azure Beastbinder — loses all abilities until your next turn",
	}.Apply(ctx)
}
