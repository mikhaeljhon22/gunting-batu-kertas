package service

import "guntingbatukertas/repo"

type PlayService struct {
	repo *repo.PlayRepo
}

func NewPlayService(repo *repo.PlayRepo) *PlayService {
	return &PlayService{
		repo: repo,
	}
}

func (s *PlayService) CreateRoom(playerName, roomName string) error {
	return s.repo.CreateRoom(playerName, roomName)
}

func (s *PlayService) JoinRoom(playerName, roomName string) error {
	return s.repo.JoinRoom(playerName, roomName)
}

func (s *PlayService) LeaveRoom(playerName, roomName string) error {
	return s.repo.LeaveRoom(playerName, roomName)
}