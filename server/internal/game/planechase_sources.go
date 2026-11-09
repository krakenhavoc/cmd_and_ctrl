package game

// planechase_sources.go makes a face-up planar card's abilities function
// from the command zone (CR 901.7, ADR 0216 §6, slice 2, #4870):
//
//	901.7. Any abilities of a face-up plane card or phenomenon card in
//	       the command zone function from that zone. The card's static
//	       abilities affect the game, its triggered abilities may
//	       trigger, and its activated abilities may be activated.
//	901.13. In multiplayer formats other than Grand Melee, plane cards
//	       and phenomenon cards are exempt from the limited range of
//	       influence option. Their abilities, and the effects of those
//	       abilities, affect all applicable objects and players in the
//	       game. (See rule 801, "Limited Range of Influence Option.")
//
// Each walk is the emblem walk's twin (emblem.go), over Game.PlanarFaceUp
// instead of each seat's emblems, so a plane needs nothing an emblem does
// not: statics join the layer pass, triggers join the harvester, the rule
// gates read it through forEachCommandSourceLocked, and its cost
// modifiers join the CR 601.2f pass. A face-up card's controller is the
// planar controller (setPlanarControllerLocked stamps it), so "you" on a
// plane is the planar controller with no special case anywhere.
//
// Activated abilities of planes (CR 901.7's last clause) are not walked:
// no plane the slices plan prints one (ADR 0216 §6). A table without
// Planechase has no face-up zone, so every walk here is an empty loop.

// planarContinuousEffectsLocked is the planar half of the layer pass's
// source gather: the static abilities of every face-up planar card,
// timestamped as the card was turned face up (CR 613.7).
//
// Caller must hold g.mu in write mode (the recompute pass does).
func (g *Game) planarContinuousEffectsLocked() []ContinuousEffect {
	if CatalogStaticAbilities == nil || g.PlanarFaceUp == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.PlanarFaceUp.Cards {
		src := &g.PlanarFaceUp.Cards[i]
		for _, ab := range staticAbilitiesOf(src) {
			out = append(out, staticContinuousEffect{ability: ab, source: src, timestamp: src.EnteredBattlefieldAt})
		}
	}
	return out
}

// forEachCommandSourceLocked calls fn with every command-zone source of
// abilities the rule gates read (ADR 0109 §5, ADR 0216 §6): every emblem
// (forEachEmblemLocked), then every face-up planar card, until fn
// returns false.
//
// Caller must hold g.mu (read or write).
func (g *Game) forEachCommandSourceLocked(fn func(src *Card) bool) {
	more := true
	g.forEachEmblemLocked(func(src *Card) bool {
		more = fn(src)
		return more
	})
	if !more || g.PlanarFaceUp == nil {
		return
	}
	for i := range g.PlanarFaceUp.Cards {
		if !fn(&g.PlanarFaceUp.Cards[i]) {
			return
		}
	}
}

// harvestFromPlanesLocked is the planar half of the trigger harvester's
// walk: the face-up cards, and, for an EventPlaneswalkFrom, the card that
// has just left them, so "whenever you planeswalk away from ~" looks back
// at it the way a leaves-the-battlefield trigger does.
//
// Caller must hold g.mu in write mode (the notify path does).
func (g *Game) harvestFromPlanesLocked(pass *harvestPass) {
	if g.Planechase == nil {
		return
	}
	if g.PlanarFaceUp != nil && len(g.PlanarFaceUp.Cards) > 0 {
		g.harvestFromZone(pass, g.PlanarFaceUp)
	}
	ev := pass.ev
	if ev.Kind != EventPlaneswalkFrom || g.PlanarDeck == nil {
		return
	}
	for _, c := range g.PlanarDeck.Cards {
		if c.InstanceID == ev.CardID {
			left := &Zone{Kind: ZoneCommand, Owner: c.Owner, Cards: []Card{c}}
			g.harvestFromZone(pass, left)
			return
		}
	}
}

// appendPlanarCostModifiersLocked adds the cost modifiers of every
// face-up planar card (ADR 0216 §6), the sibling of ADR 0140's eminence
// walk. Every modifier counts, not only those marked Eminence: all of a
// face-up plane's abilities function from the command zone (CR 901.7).
// The source is bound with the planar controller as its controller.
//
// Caller must hold g.mu.
func (g *Game) appendPlanarCostModifiersLocked(out []boundCostModifier, q CostQuery) []boundCostModifier {
	if CatalogCostModifiers == nil || g.Planechase == nil || g.PlanarFaceUp == nil {
		return out
	}
	for i := range g.PlanarFaceUp.Cards {
		mods := costModifiersOf(&g.PlanarFaceUp.Cards[i])
		if len(mods) == 0 {
			continue
		}
		src := g.PlanarFaceUp.Cards[i]
		src.Controller = g.Planechase.Controller
		for _, m := range mods {
			if !m.pricesAnnouncement(q) {
				continue
			}
			out = append(out, boundCostModifier{modifier: m, source: src})
		}
	}
	return out
}

// forEachPlanarSourceLocked calls fn with every face-up planar card until fn
// returns false: the planar half of forEachCommandSourceLocked, for the
// readers that walk the battlefield's sources themselves (cast timing).
//
// Caller must hold g.mu (read or write).
func (g *Game) forEachPlanarSourceLocked(fn func(src *Card) bool) {
	if g.PlanarFaceUp == nil {
		return
	}
	for i := range g.PlanarFaceUp.Cards {
		if !fn(&g.PlanarFaceUp.Cards[i]) {
			return
		}
	}
}
