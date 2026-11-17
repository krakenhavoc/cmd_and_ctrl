package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mishra, Lost to Phyrexia — Legendary Artifact Creature — Phyrexian
// Artificer, 9/9. The combined back face of Mishra, Claimed by Gix and
// Phyrexian Dragon Engine (CR 712.5b); never a card in a deck, only the
// melded permanent the two become (ADR 0145, #2699), here for The Brothers'
// War batch 20 (#5482):
//
//	"Whenever Mishra enters or attacks, choose three —
//	 • Target opponent discards two cards.
//	 • Mishra deals 3 damage to any target.
//	 • Destroy target artifact or planeswalker.
//	 • Creatures you control gain menace and trample until end of turn.
//	 • Creatures you don't control get -1/-1 until end of turn.
//	 • Create two tapped Powerstone tokens."
//
// One trigger with two conditions (WhenThisEntersOrAttacks); the three
// different modes are chosen as it goes on the stack and resolve in
// printed order (CR 700.2d, 608.2c). The two creature modes affect the
// creatures there are as the ability resolves (CR 611.2c). The opponent
// discards of their own choice.
//
// "It enters tapped and attacking" is the melded permanent's own entry
// replacement, because the only thing that ever puts it onto the battlefield
// is Mishra, Claimed by Gix's meld ability: the entry is tapped, and what it
// attacks is asked of its controller (CR 508.4, EntryAttack). It was never
// declared as an attacker, so its own attack trigger does not fire for that
// combat, and the enters trigger does.
func init() {
	Register(Spec{
		OracleID:     "3989308a-4c0f-493a-b38f-cb4f5019dfe7",
		Name:         "Mishra, Lost to Phyrexia",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			RowLabel:        "This creature enters tapped and attacking.",
			EntryOnly:       true,
			SelfReplacement: true,
			Watches:         []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					ev.OldZone != game.ZoneBattlefield && src != nil && ev.CardID == src.InstanceID
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			EntryAttack: &game.EntryAttackChoice{Question: "Mishra, Lost to Phyrexia — what does it attack?"},
			Label:       "Mishra, Lost to Phyrexia: enters tapped and attacking",
		}},
		Triggered: []game.TriggeredAbility{s82BroB20MishraModes()},
	})
}

// s82BroB20MishraModes is the "enters or attacks, choose three" trigger.
func s82BroB20MishraModes() game.TriggeredAbility {
	ab := WhenThisEntersOrAttacks("Mishra, Lost to Phyrexia — choose three",
		func(*game.Game, *game.StackItem) error { return nil })
	ab.Modes = ChooseN("Choose three", 3, 3,
		ModeDoing("Target opponent discards two cards.",
			TargetPlayer("target opponent", Opponent()), s82BroB20MishraDiscard),
		ModeDoing("Mishra deals 3 damage to any target.", TargetAny(), socB20DamageToTheModeTarget(3)),
		ModeDoing("Destroy target artifact or planeswalker.",
			TargetPermanent("target artifact or planeswalker", Or(Artifact(), Planeswalker())), DestroyTheModesTarget),
		ModeDoing("Creatures you control gain menace and trample until end of turn.", nil, s82BroB20MishraEvasion),
		ModeDoing("Creatures you don't control get -1/-1 until end of turn.", nil, s82BroB20MishraShrink),
		ModeDoing("Create two tapped Powerstone tokens.", nil, s82BroB20MishraPowerstones),
	)
	return ab
}

// The opponent chooses the two cards to discard (a hand of one gives up what
// it has, CR 701.9a).
func s82BroB20MishraDiscard(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: t.ID, Source: item.SourceCardID, N: 2})
	return nil
}

func s82BroB20MishraEvasion(_ *game.StackItem, ctx *Context, _ int) error {
	return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"menace", "trample"},
		Label: "Mishra, Lost to Phyrexia — menace and trample"}.Apply(ctx)
}

func s82BroB20MishraShrink(_ *game.StackItem, ctx *Context, _ int) error {
	return BoostUntilEOT{Match: And(Creature(), OpponentControls()), Power: -1, Toughness: -1,
		Label: "Mishra, Lost to Phyrexia — -1/-1"}.Apply(ctx)
}

func s82BroB20MishraPowerstones(_ *game.StackItem, ctx *Context, _ int) error {
	return CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 2}.Apply(ctx)
}
