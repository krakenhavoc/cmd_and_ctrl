package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fight the _____ Fight — Enchantment — Aura {1}{G}:
//
//	"Enchant creature you control
//	 When this Aura enters, you may put a name sticker on it. When you do,
//	 enchanted creature fights up to one target creature you don't control.
//	 Enchanted creature gets +0/+2 for each name sticker on this Aura with
//	 eight or more letters."
//
// "When you do" is a reflexive trigger (CR 603.12), created only when a
// sticker went on: it goes on the stack above the enters trigger and its
// "up to one target creature you don't control" is chosen then. The Aura
// rides the trigger's payload, so "enchanted creature" is read as the
// trigger resolves (CR 608.2): the creature the Aura enchants then. An
// Aura that has left, or enchants nothing, fights nothing.
//
// The pump counts the name stickers on the Aura itself whose words have
// eight or more letters (CR 123.6a: a letter is a letter, not a hyphen or
// an apostrophe). It is re-read every recompute, so a sticker put on later
// counts at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f589a11-1601-4a17-92d6-58e9dd9b00f2",
		Name:         "Fight the _____ Fight",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttachedPer(0, 2, nameStickersWithAtLeastLetters(8)),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fight the _____ Fight — you may put a name sticker on it. When you do, enchanted creature fights up to one target creature you don't control",
				Do(
					PutASticker{On: StickerOnThis, Kinds: []game.StickerKind{game.StickerName}},
					Instruction(fightTheBlankFightWhenYouDo),
				)),
		},
	})
}

// nameStickersWithAtLeastLetters counts the name stickers on the source
// whose word has at least n letters. It reads the source's own stickers
// only, so it is safe inside a layer pass.
func nameStickersWithAtLeastLetters(n int) func(g *game.Game, source *game.Card) int {
	return func(_ *game.Game, source *game.Card) int {
		return nameStickerCount(*source, func(text string) bool { return game.StickerLetters(text) >= n })
	}
}

// fightTheBlankFightWhenYouDo creates the reflexive trigger once a sticker
// went on.
func fightTheBlankFightWhenYouDo(ctx *Context) error {
	if _, ok := placedNameSticker(ctx); !ok {
		return nil
	}
	return ReflexiveTrigger{
		Label: "Fight the _____ Fight — enchanted creature fights up to one target creature you don't control",
		Cards: []uuid.UUID{ctx.Source()},
		Body:  fightTheBlankFightBody,
	}.Apply(ctx)
}

var fightTheBlankFightBody = game.ReflexiveBody("fight-the-blank-fight/fight", simpleBody(fightTheBlankFightFight),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("up to one target creature you don't control", Not(YouControl())).WithCount(0, 1)
	}))

func fightTheBlankFightFight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	auras := ctx.PayloadCards()
	if len(auras) == 0 {
		return nil
	}
	aura, ok := g.LookupCardForEffect(auras[0])
	if !ok || !onBattlefield(g, aura.InstanceID) {
		return nil
	}
	host := g.AttachedHostOf(&aura)
	if host == nil {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return b10Fight(ctx, host.InstanceID, t.ID)
		}
	}
	return nil
}
