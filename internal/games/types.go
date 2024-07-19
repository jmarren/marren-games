package games

type Game struct {
	GameId           int64
	GameName         string
	GameTotalMembers int64
}
type GameInvite struct {
	GameName    string
	GameId      int64
	CreatorName string
	CreatorId   int64
}

type GamesPageData struct {
	CurrentGames []Game
	GameInvites  []GameInvite
}

type ResultsData struct {
	AnswersData
	ScoreboardData []UserScore
}

type GamePlayData struct{}

type QuestionData struct{}
