package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b021_helpers.go — Secret Lair Drop, other creatures (S82 older
// sets, batch 021, #5462): the shared shapes of Maro, Dauntless Dourbark,
// Ruhan of the Fomori, Sire of Insanity and Hokori, Dust Drinker.

// S82OldB021CardsInHandPT is "<name>'s power and toughness are each equal to
// the number of cards in your hand": a layer-7a characteristic-defining
// ability (CR 604.3, 613.4a) counted live. DependsOnHandSize drops the layer
// cache when a hand changes size.
func S82OldB021CardsInHandPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:             label,
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7A_CDA,
		DependsOnHandSize: true,
		AppliesTo:         SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
				n = p.Hand.Size()
			}
			c.Power, c.Toughness = n, n
		},
	}
}

// S82OldB021ForestsPlusTreefolkPT is Dauntless Dourbark's characteristic-
// defining ability: the number of Forests you control plus the number of
// Treefolk you control (a Forest Treefolk is counted once for each).
func S82OldB021ForestsPlusTreefolkPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			// The two types are words of the source's text (CR 612.2).
			forest, treefolk := source.TextSubtype("Forest"), source.TextSubtype("Treefolk")
			n := 0
			for _, p := range g.BattlefieldCardsForEffect() {
				if p.Controller != source.Controller {
					continue
				}
				if p.HasSubtype(forest) {
					n++
				}
				if p.HasSubtype(treefolk) {
					n++
				}
			}
			c.Power, c.Toughness = n, n
		},
	}
}

// S82OldB021TrampleWhileAnotherTreefolk is "This creature has trample as long
// as you control another Treefolk."
func S82OldB021TrampleWhileAnotherTreefolk() game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.InstanceID != source.InstanceID {
			return false
		}
		treefolk := source.TextSubtype("Treefolk")
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.HasSubtype(treefolk) {
				return true
			}
		}
		return false
	}, "trample")
	ab.Label = "This creature has trample as long as you control another Treefolk."
	return ab
}

// EachPlayerDiscardsTheirHand is "each player discards their hand" as a trigger body
// (Sire of Insanity, Mindslicer):
// every player in turn order from the controller discards their whole hand
// (no choice; an empty hand discards nothing).
func EachPlayerDiscardsTheirHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range TablePlayers(ctx) {
		if _, err := DiscardWholeHand(g, id); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB021RuhanChoose is Ruhan's beginning-of-combat trigger: choose an
// opponent at random; Ruhan attacks that player this combat if able.
func S82OldB021RuhanChoose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !OnBattlefield(g, item.SourceCardID) || ctx.IsNewSourceObject(item.SourceCardID) {
		return nil
	}
	opps := ctx.Opponents()
	if len(opps) == 0 {
		return nil
	}
	pick := g.ChooseAtRandomForEffect(RandomDraw(ctx), opps, 1)
	if len(pick) == 0 {
		return nil
	}
	dur, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     []game.Mod{game.AttackPlayerIfAbleMod(pick[0])},
		Duration: dur,
		Label:    "Ruhan of the Fomori — attacks that player this combat if able",
	}.Apply(ctx)
}

// S82OldB021UntapALand is Hokori's upkeep trigger: "that player untaps a land
// they control". The player whose upkeep it is chooses (CR 608.2d); a player
// with one tapped land untaps it and one with none does nothing.
func S82OldB021UntapALand(g *game.Game, item *game.StackItem) error {
	player := TriggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	var tapped []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsLand() && c.Tapped {
			tapped = append(tapped, c.InstanceID)
		}
	}
	switch len(tapped) {
	case 0:
		return nil
	case 1:
		return UntapTarget{Target: tapped[0]}.Apply(NewContext(g, item))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Hokori, Dust Drinker — untap a land you control",
		Cards:    tapped,
		Min:      1,
		Max:      1,
		ThenKey:  s82OldB021UntapPick,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB021UntapPick untaps the land the player chose. An on-disk key:
// never renamed, never reused.
var s82OldB021UntapPick = AskedPick("choose-cards/s82-old-b021-hokori-untap-a-land",
	func(ctx *Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

// S82OldB021FourteenthDoctorReveal is The Fourteenth Doctor's cast trigger:
// reveal the top fourteen cards of your library, put all Doctor cards
// revealed this way into your graveyard and the rest on the bottom of your
// library in a random order. The Doctors are put (not milled), so each one
// carries the record "put there from your library this turn" that the
// spell's enter-as-a-copy clause reads.
func S82OldB021FourteenthDoctorReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	revealed := g.RevealTopOfLibraryForEffect(player, item.SourceCardID, 14, "The Fourteenth Doctor — reveal the top fourteen cards")
	doctor := "Doctor"
	if src, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		doctor = src.TextSubtype("Doctor") // a word of the source's text (CR 612.2)
	}
	var doctors, rest []uuid.UUID
	for _, id := range revealed {
		if c, ok := g.LookupCardForEffect(id); ok && c.HasSubtype(doctor) {
			doctors = append(doctors, id)
		} else {
			rest = append(rest, id)
		}
	}
	if err := RestIntoGraveyard(g, doctors); err != nil {
		return err
	}
	if len(rest) == 0 {
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, rest)
}

// S82OldB021DoctorCardsFromYourLibraryThisTurn are the candidates of The
// Fourteenth Doctor's "you may have it enter as a copy of a Doctor card in
// your graveyard that was put there from your library this turn".
func S82OldB021DoctorCardsFromYourLibraryThisTurn(g *game.Game, controller, self uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	doctor := "Doctor"
	if src, ok := g.LookupCardForEffect(self); ok {
		doctor = src.TextSubtype("Doctor") // a word of the entering card's text (CR 612.2)
	}
	var out []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID != self && c.IsPermanent() && c.HasSubtype(doctor) && c.PutInGraveyardFromLibraryThisTurn() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// S82OldB021GainsHasteUntilEndOfTurn is "If you do, it gains haste until end
// of turn" on an entry copy: the copy's except clause, which runs only when
// the copy was chosen, asks the entry for a haste that ends at the cleanup
// step (game.ReplacementEvent.EntersWithHasteTillEOT).
func S82OldB021GainsHasteUntilEndOfTurn(ev *game.ReplacementEvent, _ *game.PrintedValues, _ *game.Game, _ *game.Card) {
	if ev != nil {
		ev.EntersWithHasteTillEOT = true
	}
}
