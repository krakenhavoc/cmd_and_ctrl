package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b437_helpers.go — shared pieces of Avacyn Restored batch 437 (#5570).

// S82OldB437ReturnAnotherCreatureYouControl is "return another creature you
// control to its owner's hand" (Havengul Skaab): ReturnOneYouControl's prompt
// with the source left out of the candidates. "Another" is the object, never
// a name; the pick is mandatory when there is one to pick, and with none the
// ability does nothing (CR 609.3).
type S82OldB437ReturnAnotherCreatureYouControl struct {
	Question string
}

func (r S82OldB437ReturnAnotherCreatureYouControl) Apply(ctx *Context) error {
	source := ctx.Source()
	return ChoosePermanents{
		Question: r.Question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.InstanceID != source && c.IsCreature() {
					out = append(out, c.InstanceID)
				}
			}
			return out, 1, 1
		},
		Keyed: BouncedChosen,
	}.Apply(ctx)
}

// S82OldB437RandomSacrificeLoop is Tyrant of Discord's "target opponent
// chooses a permanent they control at random and sacrifices it. If a nonland
// permanent is sacrificed this way, repeat this process." The pick comes from
// the game's keyed random stream (so an undo rewinds it) and goes through the
// sacrifice's own continuation: the process repeats only once the sacrifice
// has really happened (a replacement that answers late, or a permanent that
// can't be sacrificed, ends it) and only after a nonland permanent. It stops
// when the opponent controls nothing.
func S82OldB437RandomSacrificeLoop(ctx *Context, opponent uuid.UUID) error {
	g := ctx.Game
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == opponent {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	pick := g.ChooseAtRandomForEffect(RandomDraw(ctx), ids, 1)
	if len(pick) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(pick[0])
	if !ok {
		return nil
	}
	ref, ok := game.TailRefByKey(s82OldB437RepeatKey)
	if !ok {
		return nil
	}
	t := ctx.ItemTail(ref)
	t.Carry = []uuid.UUID{opponent}
	if !c.IsLand() {
		t.N = 1
	}
	return g.SacrificeAllKeyedForEffect(ctx.Source(), pick, t)
}

// s82OldB437RepeatKey names the sacrifice's continuation: when the permanent
// really was sacrificed and was a nonland one (N = 1), the process repeats. An
// on-disk key: never renamed, never reused. The loop finds it by key because
// its body calls the function that names it.
const s82OldB437RepeatKey = "then/tyrant-of-discord-repeat"

var _ = CardTail(s82OldB437RepeatKey, func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N != 1 || len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	return S82OldB437RandomSacrificeLoop(ctx, o.Tail.Carry[0])
})

// S82OldB437XDamageToPlayerOrWalkerAndTheirCreatures is "<this> deals X damage
// to target player or planeswalker and each creature that player or that
// planeswalker's controller controls" (Lavalanche, Bonfire of the Damned): X
// rides the cast. The player is the target, or the controller of the targeted
// planeswalker, read before any damage lands; the creatures are snapshotted
// first and every point is one simultaneous damage instance (CR 608.2). With
// its one target illegal the spell does nothing (CR 608.2b).
func S82OldB437XDamageToPlayerOrWalkerAndTheirCreatures(ctx *Context) error {
	return S82OldB772DamageToPlayerOrWalkerAndTheirCreatures(ctx, ctx.X())
}

// S82OldB437CantAttackOrBlockUnlessPairedWithSoulbond is "This creature can't
// attack or block unless it's paired with a creature with soulbond"
// (Flowering Lumberknot, CR 702.95, CR 508.1c, CR 509.1b): the attack half is
// an attack restriction on the creature itself, the block half a rule that
// bars it as a blocker of any attacker. Both read the pair live (the partner
// must be on the battlefield and have soulbond through its layered
// abilities); a creature that is unpaired, or paired with one that has lost
// soulbond, does neither.
func S82OldB437CantAttackOrBlockUnlessPairedWithSoulbond() (game.StaticAbility, game.BlockRule) {
	attack := selfAttackTargetRestriction(game.AttackTargetRestriction{MustBePairedWithKeyword: game.KeywordSoulbond})
	attack.Label = "Can't attack unless it's paired with a creature with soulbond."
	self := OnSelf()
	block := game.BlockRule{
		Reason: game.BlockReasonCantBlockAttacker,
		Label:  "it can't block unless it's paired with a creature with soulbond",
		Pair: func(g *game.Game, _, blocker, source *game.Card) bool {
			if !self(g, blocker, source) {
				return false
			}
			p := g.PairedPartner(blocker)
			return p == nil || !game.HasKeyword(p, game.KeywordSoulbond)
		},
	}
	return attack, block
}
