package cli

type OutfitSession struct {
	globalShown   map[string]bool
	categoryShown map[string]map[string]bool
}

func NewOutfitSession() *OutfitSession {
	return &OutfitSession{
		globalShown:   map[string]bool{},
		categoryShown: map[string]map[string]bool{},
	}
}

func (s *OutfitSession) MarkGlobalShown(outfitKey string) {
	s.globalShown[outfitKey] = true
}

func (s *OutfitSession) MarkCategoryShown(fileName, category string) {
	if s.categoryShown[category] == nil {
		s.categoryShown[category] = map[string]bool{}
	}
	s.categoryShown[category][fileName] = true
}

func (s *OutfitSession) IsGlobalShown(outfitKey string) bool {
	return s.globalShown[outfitKey]
}

func (s *OutfitSession) IsCategoryShown(fileName, category string) bool {
	return s.categoryShown[category] != nil && s.categoryShown[category][fileName]
}

func (s *OutfitSession) ResetGlobal() {
	s.globalShown = map[string]bool{}
}

func (s *OutfitSession) ResetCategory(category string) {
	delete(s.categoryShown, category)
}

func (s *OutfitSession) ResetAll() {
	s.ResetGlobal()
	s.categoryShown = map[string]map[string]bool{}
}

func (s *OutfitSession) GlobalShownCount() int {
	return len(s.globalShown)
}

func (s *OutfitSession) CategoryShownCount(category string) int {
	return len(s.categoryShown[category])
}

func (s *OutfitSession) TrackedCategoryCount() int {
	return len(s.categoryShown)
}

func (s *OutfitSession) AddSkipped(outfitKey string) {
	s.MarkGlobalShown(outfitKey)
}

func (s *OutfitSession) AddCategorySkipped(fileName, category string) {
	s.MarkCategoryShown(fileName, category)
}

func (s *OutfitSession) IsGloballySkipped(outfitKey string) bool {
	return s.IsGlobalShown(outfitKey)
}

func (s *OutfitSession) IsCategorySkipped(fileName, category string) bool {
	return s.IsCategoryShown(fileName, category)
}

func (s *OutfitSession) GlobalSkippedCount() int {
	return s.GlobalShownCount()
}

func (s *OutfitSession) CategorySkippedCount(category string) int {
	return s.CategoryShownCount(category)
}
