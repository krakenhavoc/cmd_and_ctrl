package game

// boast.go — #2697, CR 702.142: "Boast — [cost]: [effect]. (Activate
// only if this creature attacked this turn and only once each turn.)"
//
// # Two rules in one keyword, and one function that holds both
//
// A boast ability is an ordinary CR 602 activated ability with two
// activation instructions printed on it (CR 602.1b), and both were
// already answerable before this file existed: TurnTally.Attacks knows
// which creatures attacked (AttackedThisTurn) and Activations.Turn
// counts the announcements of each ability (ActivatedThisTurn,
// activation_tally.go, which names boast as a consumer). What was
// missing was ONE place that joins them, so that a card file does not
// assemble "attacked this turn" and "once each turn" by hand and so
// that the engine, the bot enumerator and the view cannot disagree
// about the join. BoastBlockLocked is that place, exactly as
// AbilityExhausted is the exhaust gate: ActivateCatalogAbility refuses
// on it, internal/legal drops the move on it and protocol's
// ActivatedAbilityView stamps `boast_blocked` from it.
//
// It answers WHICH half failed rather than a bare bool, because the
// two recover differently and a player is owed the reason: a creature
// that has not attacked can still boast later this turn; one that has
// spent its boast cannot until the next turn.
//
// # Why the limit is a modifier and not a literal 1
//
// Birgi, God of Storytelling prints "Creatures you control can boast
// twice during each of your turns rather than once". That is a
// per-controller change to the "only once each turn" half, so the
// limit is read through BoastLimit statics on the battlefield and the
// largest applicable limit wins (the printed default being 1). It is
// deliberately a MAXIMUM and not a sum: two Birgis do not let a
// creature boast three times, because each says "twice ... rather than
// once" — a replacement of the number, not an addition to it.
//
// The limit is read at the moment of activation. If Birgi leaves after
// a creature has boasted twice, the creature is simply spent (two
// activations against a limit of one); if Birgi leaves after one, it is
// spent as well. Nothing is stored that could outlive the static
// ability, which is the duration (the same property the exhaust
// permission has).
//
// # What counts as an activation
//
// The announcement, not the resolution — the tally is written at the
// announce on the one path that pays for an activation. A boast that is
// countered on the stack or fizzles for want of a target is still
// spent, which is the printed "only once each turn" and the #259
// direction the other way would be a second free use.
//
// Per ability, not per creature: a creature with two boast abilities
// may activate each once. That is the tally's key (object, label), and
// it is what the rule says: the clause is part of each ability.

// BoastBlock is why a boast ability cannot be activated right now.
type BoastBlock int

const (
	// BoastClear: nothing about the boast clause objects. Also the
	// answer for an ability that is not a boast ability at all.
	BoastClear BoastBlock = iota
	// BoastNotAttacked: the creature has not attacked this turn.
	BoastNotAttacked
	// BoastSpent: the ability has been activated as many times this
	// turn as the creature may.
	BoastSpent
)

// String is the stable wire token for the view's `boast_blocked`.
func (b BoastBlock) String() string {
	switch b {
	case BoastNotAttacked:
		return "not_attacked"
	case BoastSpent:
		return "used"
	default:
		return ""
	}
}

// BoastLimit is one "creatures can boast N times each turn" static
// contributed by a permanent on the battlefield (Birgi, God of
// Storytelling). A struct of hooks, for the reason StaticAbility,
// CostModifier and ExhaustPermission are: a card file writes a literal.
type BoastLimit struct {
	// Label is the clause as printed, for debugging a creature that
	// boasted a number of times nobody expected.
	Label string

	// Limit is how many times each boast ability may be activated each
	// turn while this static applies. The printed default is 1, so a
	// value below 2 changes nothing and is refused at Register.
	Limit int

	// Applies decides whether this permanent's static reaches the
	// creature that wants to boast. `grantor` is the permanent
	// contributing the static, so grantor.Controller is the "you" of the
	// ability. Birgi: boaster.Controller == grantor.Controller and it is
	// that player's turn. Nil is refused at Register: it would reach
	// every creature of every player, always, which no card prints.
	//
	// Evaluated under g.mu. READ-ONLY: a *ForEffect accessor is fine, a
	// mutator is a bug and a public locking one is a deadlock.
	Applies func(g *Game, boaster Card, grantor Card) bool

	// ActiveWhen is the ADR 0071 designation gate, the same slot
	// ExhaustPermission carries.
	ActiveWhen Designation
}

// CatalogBoastLimits is the catalog hook, mirroring
// CatalogExhaustPermissions. Nil, or a nil return, means the card
// changes nobody's boast limit — which is every card but one.
var CatalogBoastLimits func(key string) []BoastLimit

// boastLimitsOf is the limits a permanent contributes right now,
// ability removal and designation gate applied. CatalogAbilityKey and
// not CatalogKey: this is a static ability, so a permanent under a
// CR 613.1f ability-removing effect has stopped granting it (ADR 0046).
func boastLimitsOf(c *Card) []BoastLimit {
	if CatalogBoastLimits == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogBoastLimits(key), func(l BoastLimit) Designation {
		return l.ActiveWhen
	})
}

// BoastLimitFor is how many times each boast ability of `boaster` may
// be activated this turn: the printed once, or the largest limit any
// battlefield static grants it (CR 113.6 — a static works from the
// battlefield, which is why this is a battlefield walk).
//
// Caller must hold g.mu.
func (g *Game) BoastLimitFor(boaster Card) int {
	limit := 1
	if g == nil || g.Battlefield == nil || CatalogBoastLimits == nil {
		return limit
	}
	for i := range g.Battlefield.Cards {
		grantor := &g.Battlefield.Cards[i]
		for _, l := range boastLimitsOf(grantor) {
			if l.Limit > limit && l.Applies != nil && l.Applies(g, boaster, *grantor) {
				limit = l.Limit
			}
		}
	}
	return limit
}

// BoastBlockLocked is the whole boast gate: BoastClear when `ab` is
// not a boast ability or nothing objects, otherwise the half of the
// printed clause that fails. Attack first, because it is the half that
// explains the other: a creature that never attacked has nothing to
// have spent.
//
// Caller must hold g.mu.
func (g *Game) BoastBlockLocked(source *Card, ab ActivatedAbilityShape) BoastBlock {
	if !ab.Boast || source == nil {
		return BoastClear
	}
	if !g.AttackedThisTurn(source.InstanceID) {
		return BoastNotAttacked
	}
	// The walk for a modifier happens only once the ability has been
	// used at all: an unused boast is clear whatever the limit is.
	used := g.ActivatedThisTurn(source.InstanceID, ab.Label)
	if used == 0 {
		return BoastClear
	}
	if used >= g.BoastLimitFor(*source) {
		return BoastSpent
	}
	return BoastClear
}

// boastErr maps a block to the engine's refusal, nil for BoastClear.
func boastErr(b BoastBlock) error {
	switch b {
	case BoastNotAttacked:
		return ErrBoastNotAttacked
	case BoastSpent:
		return ErrBoastSpent
	default:
		return nil
	}
}
