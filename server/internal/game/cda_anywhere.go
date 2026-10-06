package game

// cda_anywhere.go — a card's power and toughness where no layer pass
// runs (#2115).
//
// The layer engine applies a characteristic-defining ability (CR 604.3)
// to permanents only. Off the battlefield a card reads its printed
// values (effectiveOf), and a `*` arrives from the importer as a 0
// stand-in (Card.VariableToughness). But CR 113.6a: "Characteristic-
// defining abilities function everywhere, even outside the game", and
// CR 208.2a says the same of a `*` power or toughness. So an effect
// that reads a card in a hand or a graveyard — Talara's Bane's "You
// gain life equal to that creature card's toughness", whose 2008-08-01
// ruling is exactly this — has to run the card's own layer-7a
// abilities itself.

// ToughnessAnywhereForEffect is the card's toughness wherever it is:
// on the battlefield its layered toughness, anywhere else its printed
// toughness with its own characteristic-defining abilities applied
// (CR 113.6a, 208.2a, 604.3) and any counters it carries. c may be a
// last-known copy; it is not written.
//
// Caller must hold g.mu.
func (g *Game) ToughnessAnywhereForEffect(c Card) int {
	if c.effective != nil {
		// A permanent: the layer pass already applied its CDA, and
		// applying it again would overwrite what later sublayers did.
		return c.CurrentToughness()
	}
	ch := effectiveOf(&c)
	if CatalogStaticAbilities != nil {
		for _, s := range CatalogStaticAbilities(catalogAbilityKeyOf(&c)) {
			if s.Layer != Layer7PT || s.SubLayer != SubLayer7A_CDA || s.Apply == nil {
				continue
			}
			if s.AppliesTo != nil && !s.AppliesTo(&c, g, &c) {
				continue
			}
			s.Apply(&ch, &c, g, &c)
		}
	}
	_, t := ptWithCounters(ch, c.Counters)
	return t
}
