package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b451_helpers.go — Magic 2012 (S82 older sets, batch 451, #5573):
// the shared shapes of Throne of Empires, Crown of Empires and Vengeful
// Pharaoh.

// S82OldB451ControlsArtifactsNamed is "you control artifacts named <a> and
// <b>": one artifact permanent of each name, read through the layered view.
func S82OldB451ControlsArtifactsNamed(g *game.Game, you uuid.UUID, names ...string) bool {
	for _, name := range names {
		found := false
		for _, p := range g.BattlefieldCardsForEffect() {
			if p.Controller == you && p.IsArtifact() && game.HasName(p, name) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// S82OldB451ThroneOfEmpires is Throne of Empires' ability: a 1/1 white
// Soldier, or five of them if you control artifacts named Crown of Empires
// and Scepter of Empires, checked as the ability resolves.
func S82OldB451ThroneOfEmpires(g *game.Game, item *game.StackItem) error {
	n := 1
	if S82OldB451ControlsArtifactsNamed(g, item.Controller, "Crown of Empires", "Scepter of Empires") {
		n = 5
	}
	return CreateToken{Template: TokenCard("1/1 white Soldier"), N: n}.Apply(NewContext(g, item))
}

// S82OldB451CrownOfEmpires is Crown of Empires' ability: tap the target
// creature, or gain control of it instead if you control artifacts named
// Scepter of Empires and Throne of Empires.
func S82OldB451CrownOfEmpires(item *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	if S82OldB451ControlsArtifactsNamed(ctx.Game, item.Controller, "Scepter of Empires", "Throne of Empires") {
		return GainControl{
			Target:     t.ID,
			Controller: item.Controller,
			Duration:   game.IndefiniteDuration(),
			Label:      "Crown of Empires — gain control of that creature",
		}.Apply(ctx)
	}
	return TapTarget{Target: t.ID}.Apply(ctx)
}

// S82OldB451DamageToYouOrYourPlaneswalker is "combat damage is dealt to you or
// a planeswalker you control".
func S82OldB451DamageToYouOrYourPlaneswalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if ev.Target == source.Controller {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsPlaneswalker() && c.Controller == source.Controller
}

// S82OldB451PharaohRises is Vengeful Pharaoh's graveyard trigger: destroy the
// target attacking creature, then put this card on top of your library.
func S82OldB451PharaohRises(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardForEffect(id, game.ZoneLibrary)
}

// S82OldB451CrownOfEmpiresAbility is S82OldB451CrownOfEmpires as an
// activated ability's Effect.
func S82OldB451CrownOfEmpiresAbility(g *game.Game, item *game.StackItem) error {
	return S82OldB451CrownOfEmpires(item, NewContext(g, item))
}

// S82OldB451PTWords words a power and toughness as "4/4".
func S82OldB451PTWords(power, toughness int) string {
	return strconv.Itoa(power) + "/" + strconv.Itoa(toughness)
}

// S82OldB451GainsWords words the keywords a mode grants: " and gains trample".
func S82OldB451GainsWords(keywords []string) string {
	if len(keywords) == 0 {
		return ""
	}
	return " and gains " + strings.Join(keywords, " and ")
}
