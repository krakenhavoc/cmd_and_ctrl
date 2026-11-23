package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b22_helpers.go — shared bodies for the S82 SNC batch 22 cards.

// s82SNC22CreaturesEnterTapped is "Creatures enter tapped" (False Floor):
// Kismet's CR 614 replacement narrowed to creatures and applying to every
// player's, the source's controller included.
func S82SNC22CreaturesEnterTapped(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature()
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// s82SNC22AllCreaturesAttackHostController is "All creatures attack
// enchanted creature's controller each combat if able" (Public Enemy): one
// CR 508.1d requirement per creature, naming the player who controls the
// enchanted creature as the combat is declared. A creature that player
// controls cannot attack them, so the requirement is impossible and
// ignored for it (CR 508.1d).
func S82SNC22AllCreaturesAttackHostController() game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		Label: "All creatures attack enchanted creature's controller each combat if able.",
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && source != nil && source.AttachedTo.Kind == game.TargetCard
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			for i := range g.Battlefield.Cards {
				host := &g.Battlefield.Cards[i]
				if host.InstanceID != source.AttachedTo.ID {
					continue
				}
				r := requirementFrom(source, uuid.Nil)
				r.MustAttack = host.Controller
				c.AttackRequirements = append(c.AttackRequirements, r)
				return
			}
		},
	}
}

// s82SNC22ContestedKind is the counter Turf War puts on lands. An on-disk
// identity: never renamed.
const S82SNC22ContestedKind = "contested"

// s82SNC22ContestedLands is the lands `player` controls that carry a
// contested counter.
func s82SNC22ContestedLands(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsLand() && c.Counters[S82SNC22ContestedKind] > 0 {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// s82SNC22SeatLands is "target land that player controls" for the seat k
// places after the caster in turn order (the table has at most four seats).
func S82SNC22SeatLands() *game.TargetSpec {
	var clauses []*game.TargetSpec
	for k := 0; k < game.MaxPlayers; k++ {
		k := k
		clauses = append(clauses, TargetPermanent("target land (player "+string(rune('1'+k))+" from you) controls", Land(),
			func(g *game.Game, caster uuid.UUID, c game.Card) bool {
				return cmdr2S38SeatFrom(g, caster, k) == c.Controller
			}))
	}
	return Clauses(clauses[0], clauses[1:]...)
}

// s82SNC22TurfWarHit is the intervening-"if" half of Turf War's second
// ability (CR 603.4): a creature dealt combat damage to a player who
// controls a land with a contested counter.
func S82SNC22TurfWarHit(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	if !ok || !dealer.IsCreature() {
		return false
	}
	return len(s82SNC22ContestedLands(g, ev.Target)) > 0
}

// s82SNC22TurfWarTake: the creature's controller gains control of one of
// the damaged player's contested lands, of their choice, and untaps it.
func S82SNC22TurfWarTake(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	var thief uuid.UUID
	if dealer, ok := g.LookupCardForEffect(ev.Source); ok {
		thief = dealer.Controller
	} else if info, ok := g.LastKnownPermanentForEffect(ev.Source); ok {
		thief = info.Controller
	}
	if thief == uuid.Nil || g.PlayerByIDForEffect(thief) == nil {
		return nil
	}
	lands := s82SNC22ContestedLands(g, ev.Target)
	switch len(lands) {
	case 0:
		return nil
	case 1:
		return s82SNC22TurfWarSteal(g, item.SourceCardID, thief, lands[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  thief,
		Source:   item.SourceCardID,
		Question: "Turf War — choose a land with a contested counter to gain control of and untap",
		Cards:    lands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  s82SNC22TurfWarPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

func s82SNC22TurfWarSteal(g *game.Game, source, thief, land uuid.UUID) error {
	if !g.GainControlForEffect(source, land, thief, game.IndefiniteDuration(), "Turf War — gain control of the land") {
		return nil
	}
	return g.UntapTargetForEffect(land)
}

// s82SNC22TurfWarPick is the picked land's theft, keyed. An on-disk key:
// never renamed, never reused.
var s82SNC22TurfWarPick = AskedPick("choose-cards/s82-snc-turf-war-take",
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 1 {
			return nil
		}
		return s82SNC22TurfWarSteal(ctx.Game, r.Source, r.Chooser, r.Picked[0])
	})

// s82SNC22CreaturesControlledBy counts the creatures `player` controls.
func S82SNC22CreaturesControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsCreature() {
			n++
		}
	}
	return n
}

// s82SNC22EntersNotAToken is Life of the Party's "When this creature
// enters, if it's not a token" (CR 603.4): the entering permanent is the
// source and is not a token.
func S82SNC22EntersNotAToken(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventETB && source != nil && ev.CardID == source.InstanceID && !source.IsToken()
}

// s82SNC22LifeCopyForNext makes the next opponent's goaded token copy, then
// (from the keyed tail) the ones after it. Each token is goaded for the
// rest of the game by the effect's controller as it lands.
func S82SNC22LifeCopyForNext(ctx *Context, players []uuid.UUID, template game.Card) error {
	for len(players) > 0 && ctx.Game.PlayerByIDForEffect(players[0]) == nil {
		players = players[1:]
	}
	if len(players) == 0 {
		return nil
	}
	player, rest := players[0], append([]uuid.UUID(nil), players[1:]...)
	tail := ctx.ItemTail(s82SNC22GoadThenNextCopy)
	tail.Carry = rest
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: player,
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: CopiedToken(template), Count: 1}},
	}, tail)
}

// s82SNC22GoadThenNextCopy goads the copy the last creation made, then makes
// the next opponent's. The next copy is rebuilt from the source as it last
// existed. An on-disk key: never renamed, never reused.
var s82SNC22GoadThenNextCopy game.TailRef

func init() {
	s82SNC22GoadThenNextCopy = CardTail("s82-snc/life-of-the-party-goad-then-next", func(ctx *Context, o game.TailOutcome) error {
		for _, id := range o.IDs {
			ctx.Game.GoadForRestOfGameForEffect(id, o.Tail.Player)
		}
		tmpl, ok := TokenCopyTemplate(ctx.Game, ctx.Source())
		if !ok {
			return nil
		}
		return S82SNC22LifeCopyForNext(ctx, o.Tail.Carry, tmpl)
	})
}

// s82SNC22IsJinnieToken reports whether a template is already one of the two
// tokens Jinnie Fay makes: the 2/2 green Cat with haste or the 3/1 green Dog
// with vigilance.
func s82SNC22IsJinnieToken(t game.Card) bool {
	for _, key := range []string{"2/2 green Cat with haste", "3/1 green Dog with vigilance"} {
		want := TokenCard(key)
		if t.Name == want.Name && t.Power == want.Power && t.Toughness == want.Toughness &&
			len(t.Keywords) == 1 && t.Keywords[0] == want.Keywords[0] && len(t.Colors) == 1 && t.Colors[0] == "G" && t.TypeLine == want.TypeLine {
			return true
		}
	}
	return false
}

// s82SNC22JinnieInstead is one of Jinnie Fay's two "you may instead create
// that many <tokens>" replacements. Jinnie prints ONE effect with two
// outcomes, so each of the pair stops applying once the instruction's
// groups are already her tokens: whichever the player answers "yes" to is
// the only one that applies (CR 614.5). The player is asked about each in
// turn, and "no" to both leaves the instruction alone.
func S82SNC22JinnieInstead(label, question, tokenKey string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:  []game.EventKind{game.EventTokenCreated},
		Optional: true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventCreateTokens || ev.TokenCount() <= 0 || ev.TokenController != src.Controller {
				return false
			}
			return ev.TokenTemplatesMatch(func(t game.Card) bool { return !s82SNC22IsJinnieToken(t) })
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.ReplaceTokenKindsWhere(func(t game.Card) bool { return !s82SNC22IsJinnieToken(t) }, TokenCard(tokenKey))
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label:          label,
		PromptQuestion: question,
	}
}
