package game

// TokenCopyTemplateForObjectForEffect copies one battlefield object's
// copiable values (CR 707.2), using last-known information once that
// object has departed (CR 608.2h). A card that has returned with a new
// object epoch is not the object the resolving trigger named.
//
// Caller must hold g.mu in write mode.
func (g *Game) TokenCopyTemplateForObjectForEffect(ref ObjectRef) (Card, bool) {
	info, ok := g.PermanentForEffect(ref)
	if !ok || info.Copiable == nil {
		return Card{}, false
	}
	values := info.Copiable.Clone()
	values.MakeToken()
	var token Card
	token.setPrintedValues(values)
	return token, true
}
