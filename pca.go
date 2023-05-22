package main

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

func ArtistTracksandFeatures(client *spotify.Client, ctx context.Context, artistID spotify.ID) (artistTracks []spotify.FullTrack, artistAudioFeatures []*spotify.AudioFeatures, err error) {
	artistTracks, err = client.GetArtistsTopTracks(ctx, spotify.ID(artistID), "JP")
	if err != nil {
		return
	}
	var artistTrackIDs []spotify.ID
	for _, track := range artistTracks {
		artistTrackIDs = append(artistTrackIDs, track.ID)
	}
	artistAudioFeatures, err = client.GetAudioFeatures(ctx, artistTrackIDs...)
	return
}
