package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB504ExileTheFirstTarget is "exile target card from a graveyard" as the
// whole of an optional trigger's effect: the first target still legal as it
// resolves (CR 608.2b) goes to exile, nothing when none is.
func S82OldB504ExileTheFirstTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if t, ok := FirstLegalTarget(ctx); ok {
		return ExileTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}

// putChosenOnTopRef is the chosen creature going on top of its owner's library,
// keyed like BouncedChosen. An on-disk key.
var putChosenOnTopRef = RegisterPermanentsPick("prompt-run/put-chosen-permanent-on-top-of-library", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if OnBattlefield(ctx.Game, id) {
				if err := ctx.Game.TuckToLibraryForEffect(id, false); err != nil {
					return err
				}
			}
		}
		return nil
	},
})

// S82OldB504OneCreatureYouControl is the offer of "a creature you control" as one
// seat's candidates: every creature that seat controls, exactly one to pick.
func S82OldB504OneCreatureYouControl(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	return CreaturesControlledByPlayer(g, of), 1, 1
}

// S82OldB504PutACreatureYouControlOnTop is "put a creature you control on top
// of its owner's library" (Nulltread Gargantuan): ReturnOneYouControl's
// resolution-time pick over your own creatures (no target; the source itself is
// a candidate), mandatory with at least one candidate and nothing with none.
func S82OldB504PutACreatureYouControlOnTop(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return ChoosePermanents{
			Question:   question,
			Candidates: S82OldB504OneCreatureYouControl,
			Keyed:      putChosenOnTopRef,
		}.Apply(ctx)
	}
}

// S82OldB504ShuffleThisIntoLibrary is "this creature's owner shuffles it into
// their library" as a battlefield trigger's effect (Blitz Hellion): the ordinary
// tuck, so a commander's owner is offered the command zone (CR 903.9), then the
// owner shuffles. It does nothing once the Hellion has left or come back as a
// new object (CR 400.7).
func S82OldB504ShuffleThisIntoLibrary(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	owner := item.Controller
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && c.Owner != uuid.Nil {
		owner = c.Owner
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{item.SourceCardID}, game.TuckOptions{},
		game.KeyedTail{Ref: ThenShuffle, Player: owner})
}

// S82OldB504ActivePlayerSacrificesAMonocoloredCreature is Defiler of Souls'
// "that player sacrifices a monocolored creature of their choice": the player
// whose upkeep it is chooses among their own monocolored creatures (exactly one
// colour, read through the layers); with none nothing happens.
func S82OldB504ActivePlayerSacrificesAMonocoloredCreature(g *game.Game, item *game.StackItem) error {
	active := g.Seats[g.Turn.ActiveSeat]
	if active == nil {
		return nil
	}
	mono := B41Monocolored()
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == active.ID && c.IsCreature() && mono(g, active.ID, c) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	return SacrificeChoice{
		Player:     active.ID,
		Candidates: candidates,
		Question:   "Defiler of Souls — sacrifice a monocolored creature",
	}.Apply(NewContext(g, item))
}

// S82OldB504EquippedBlocksItDoesntUntap is "Whenever equipped creature blocks a
// creature, that creature doesn't untap during its controller's next untap
// step" (Shield of the Righteous): Labyrinth Minotaur's per-pair trigger with the
// blocker being the creature the Equipment is attached to (the trigger, its
// effect and its fill-in are WheneverThisBlocksItDoesntUntap's; only the
// condition is replaced). "That creature" is
// the attacker, recorded as the trigger goes on the stack with its CR 400.7
// epoch.
func S82OldB504EquippedBlocksItDoesntUntap(label string) game.TriggeredAbility {
	t := WheneverThisBlocksItDoesntUntap(label)
	t.AppliesTo = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Target != uuid.Nil && source.IsAttachedTo(ev.CardID)
	}
	return t
}

// S82OldB504UnscytheDamagedCreatureDied is the condition of Unscythe's trigger: a
// creature died that the Equipment's equipped creature had dealt damage to this
// turn (each read as the permanent it is now or most recently was, CR 400.7).
func S82OldB504UnscytheDamagedCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if _, ok := DiedCreature(ev, g); !ok {
		return false
	}
	host := g.AttachedHostOf(source)
	return host != nil && host.InstanceID != ev.CardID && DealtDamageByThisTurn(g, host.InstanceID, ev.CardID)
}

// s82OldB504UnscytheZombie is the rest of Unscythe's trigger: once the card has
// reached exile, create a 2/2 black Zombie. An on-disk key.
var s82OldB504UnscytheZombie = CardTail("then/unscythe-killer-of-kings-zombie", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
})

// S82OldB504UnscytheExile is "you may exile that card. If you do, create a 2/2
// black Zombie creature token": the card that died is exiled from the graveyard,
// and the Zombie follows once it has.
func S82OldB504UnscytheExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	ctx := NewContext(g, item)
	return ExileTarget{Target: id, ThenTail: ctx.ItemTail(s82OldB504UnscytheZombie)}.Apply(ctx)
}
