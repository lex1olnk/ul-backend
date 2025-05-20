package queries

var GetAverageMapStatsQuery = `
SELECT
    m.map_name,
    COUNT(*) AS matches,
    SUM(CASE WHEN mp.is_winner = true THEN 1 ELSE 0 END) AS wins,
	CAST(SUM(mp.rating) / COUNT(*) as DECIMAL(10, 2)) as avg_rating,
    ROUND(SUM(CASE WHEN mp.is_winner = true THEN 1 ELSE 0 END)::DECIMAL / COUNT(*)::DECIMAL * 100, 2) as winrate 
FROM 
    match_players mp
JOIN 
    maps m ON mp.map_id = m.map_id
WHERE 
    mp.player_id = $1  -- Замени на нужный player_id
GROUP BY 
    m.map_name
ORDER BY 
    matches DESC;
`

var GetAverageStatsQuery = `
WITH player_stats AS (
    SELECT 
        mp.player_id,
        p.nickname,
        p.img,
		p.faceit,
        p.ul_rating,
        COUNT(p.nickname) AS maps,
        -- Основные метрики
        SUM(kills) AS kills,
        SUM(deaths) AS deaths,
        SUM(assists) AS assists,
        SUM(firstkills) AS fk,
        SUM(firstdeaths) AS fd,
        SUM(flash_assists) AS flashes,
        SUM(grenades_damages) AS nades,
        ROUND(SUM(impact)::DECIMAL, 2) AS impact,
        SUM(exchanged) AS exchanged,
        -- Статистика клатчей по типам
        SUM(clutches[1]) AS clutches_1v1,
        SUM(clutches[2]) AS clutches_1v2,
        SUM(clutches[3]) AS clutches_1v3,
        SUM(clutches[4]) AS clutches_1v4,
        SUM(clutches[5]) AS clutches_1v5,
        -- Остальные метрики
        ROUND(SUM(kastscore)::DECIMAL / NULLIF(SUM(mp.rounds)::DECIMAL, 0) * 100, 2) AS kast,
        SUM(CASE mp.is_winner WHEN true THEN 1 ELSE 0 END)::DECIMAL /
        NULLIF(COUNT(mp.player_id)::DECIMAL, 0) * 100 AS winrate,
        SUM(kills)::DECIMAL / NULLIF(SUM(deaths)::DECIMAL, 0) AS kd_ratio,
        COALESCE(SUM(headshots), 0) / NULLIF(SUM(mp.kills)::DECIMAL, 0) * 100 AS total_hs_ratio,
        COALESCE(SUM(mp.damage), 0) / NULLIF(SUM(mp.rounds)::DECIMAL, 0) AS total_avg,
        COALESCE(SUM(mp.rating), 0) / NULLIF(COUNT(mp.match_id)::DECIMAL, 0) AS avg_rating
    FROM match_players mp
    JOIN players p ON mp.player_id = p.player_id
    LEFT JOIN matches m ON mp.match_id = m.match_id
    GROUP BY mp.player_id, p.nickname, p.ul_rating, p.img, p.faceit
),
target_player AS (
    SELECT * FROM player_stats WHERE player_id = $1
),
comparison AS (
    SELECT
        tp.player_id,
        tp.nickname,
		    tp.faceit,
        tp.ul_rating,
        tp.img,
        tp.kills,
        tp.deaths,
        tp.assists,
        tp.fk,
        tp.fd,
        tp.flashes,
        tp.nades,
        tp.impact,
        tp.kast,
        tp.maps,
        tp.exchanged,
        -- Статистика клатчей
        tp.clutches_1v1,
        tp.clutches_1v2,
        tp.clutches_1v3,
        tp.clutches_1v4,
        tp.clutches_1v5,
        ROUND(tp.winrate, 2) AS winrate,
        -- Процентили
        (1 - ROUND(COUNT(CASE WHEN ps.kd_ratio <= tp.kd_ratio THEN 1 END)::DECIMAL / 
                  NULLIF(COUNT(tp.player_id)::DECIMAL, 0), 2)) * 100 AS kd_percentile,
        (1 - ROUND(COUNT(CASE WHEN ps.total_hs_ratio <= tp.total_hs_ratio THEN 1 END)::DECIMAL / 
                  NULLIF(COUNT(tp.player_id)::DECIMAL, 0), 2)) * 100 AS hs_percentile,
        (1 - ROUND(COUNT(CASE WHEN ps.total_avg <= tp.total_avg THEN 1 END)::DECIMAL / 
                  NULLIF(COUNT(tp.player_id)::DECIMAL, 0), 2)) * 100 AS avg_percentile,
        (1 - ROUND(COUNT(CASE WHEN ps.avg_rating <= tp.avg_rating THEN 1 END)::DECIMAL / 
                  NULLIF(COUNT(tp.player_id)::DECIMAL, 0), 2)) * 100 AS rating_percentile
    FROM target_player tp
    CROSS JOIN player_stats ps
    GROUP BY 
        tp.player_id, tp.nickname, tp.ul_rating, tp.img, 
        tp.kills, tp.deaths, tp.assists, tp.kast, tp.fk, tp.fd, 
        tp.flashes, tp.nades, tp.impact, tp.maps, tp.winrate,
        tp.exchanged, tp.faceit,
        tp.clutches_1v1, tp.clutches_1v2, tp.clutches_1v3, 
        tp.clutches_1v4, tp.clutches_1v5
)
SELECT 
    c.*,
    ROUND(tp.kd_ratio, 2) AS kd,
    ROUND(tp.total_hs_ratio, 2) AS hs_percent,
    ROUND(tp.total_avg, 2) AS avg_damage,
    ROUND(tp.avg_rating::DECIMAL, 2) AS rating
FROM comparison c
JOIN target_player tp ON c.player_id = tp.player_id;
`

var GetPlayerStats = `SELECT 
	p.player_id,
	p.nickname,
	p.UL_rating,
	p.img,
	m.ul_tournament_id,
	COUNT(mp.match_id) AS matches,
	COALESCE(SUM(mp.kills), 0) AS kills,
	COALESCE(SUM(mp.deaths), 0) AS deaths, 
	COALESCE(SUM(mp.assists), 0) AS assists,
	COALESCE(SUM(mp.headshots), 0) AS headshots,
	COALESCE(SUM(mp.kastscore), 0) AS kastscore,
	COALESCE(SUM(mp.firstkills), 0) AS firstKills,
	COALESCE(SUM(mp.firstDeaths), 0) AS firstDeaths,
	SUM(clutches[1]) AS clutches_1v1,
	SUM(clutches[2]) AS clutches_1v2,
	SUM(clutches[3]) AS clutches_1v3,
	SUM(clutches[4]) AS clutches_1v4,
	SUM(clutches[5]) AS clutches_1v5,
	COALESCE(SUM(mp.damage), 0) AS damage,
	COALESCE(SUM(mp.rating) / COUNT(mp.match_id), 0) AS rating,
	CAST(COALESCE(SUM(mp.impact) / COUNT(mp.match_id), 0) as DECIMAL(10, 2)) AS impact,
	COALESCE(SUM(mp.rounds), 0) AS total_rounds
	FROM players p
	LEFT JOIN match_players mp ON p.player_id = mp.player_id
	LEFT JOIN matches m ON mp.match_id = m.match_id
	GROUP BY p.player_id, p.nickname, p.UL_rating, m.ul_tournament_id
	ORDER BY p.nickname`

var PlayerUlTournamentsQuery = `
	SELECT 
		m.ul_tournament_id, ul.name
	FROM 
		match_players mp
	LEFT JOIN
		matches m on mp.match_id = m.match_id
	LEFT JOIN 
		ul_tournaments ul on ul.id = m.ul_tournament_id
	WHERE 
		player_id = $1
	GROUP BY 
		m.ul_tournament_id, ul.name
`

var PlayerMatchesByUlIdQuery = `
SELECT 
	mp.match_id,
	mp.kills,
	mp.deaths,
	mp.assists,
	mp.finished_at,
	mp.rating,
	mp.is_winner,
	maps.map_name
FROM 
	match_players mp
LEFT JOIN
	matches m on mp.match_id = m.match_id
LEFT JOIN 
	maps ON maps.map_id = mp.map_id
WHERE 
	player_id = $1    
	AND (
	-- Используем IS NOT DISTINCT FROM для корректного сравнения NULL
		m.ul_tournament_id IS NOT DISTINCT FROM $2::uuid
	)
ORDER BY
	mp.finished_at DESC
LIMIT 
	10
OFFSET 
	$3
`

var MatchDamageQuery = `
	query GetMatchDamages($matchId: Int!) {
		damages: match_damages(where: {match_id: {_eq: $matchId}}) {
			roundId: round_id
			inflictorId: inflictor_id
			victimId: victim_id
			weaponId: weapon_id
			hitboxGroup: hitbox_group
			damageReal: damage_real
			damageNormalized: damage_normalized
			hits
			__typename
		}
	}
`

var MapMatchKillsQuery = `
	query GetMatchKills($matchId: Int!) {
		kills: match_kills(
			where: {match_id: {_eq: $matchId}}
			order_by: {created_at: asc}
		) {
			roundId: round_id
			createdAt: created_at
			killerId: killer_id
			victimId: victim_id
			assistantId: assistant_id
			weaponId: weapon_id
			isHeadshot: is_headshot
			isWallbang: is_wallbang
			isOneshot: is_oneshot
			isAirshot: is_airshot
			isNoscope: is_noscope
			killerPositionX: killer_position_x
			killerPositionY: killer_position_y
			victimPositionX: victim_position_x
			victimPositionY: victim_position_y
			killerBlindedBy: killer_blinded_by
			killerBlindDuration: killer_blind_duration
			victimBlindedBy: victim_blinded_by
			victimBlindDuration: victim_blind_duration
			isTeamkill: is_teamkill
			__typename
		}
	}
`

var MatchCluchesQuery = `query GetMatchClutches($matchId: Int!) {
  clutches: match_clutches(where: {match_id: {_eq: $matchId}}) {
    roundId: round_id
    userId: user_id
    createdAt: created_at
    success
    amount
    __typename
  }
}
  `

var FullMatchQuery = `query GetMatchStats($matchId: Int!, $gameId: smallint!) {
	match: matches_by_pk(id: $matchId) {
	  id
	  type
	  status
	  bestOf: best_of
	  gameId: game_id
	  hasWinner: has_winner
	  startedAt: started_at
	  finishedAt: finished_at
	  maxRoundsCount: max_rounds_count
	  serverInstanceId: server_instance_id
	  cancellationReason: cancellation_reason
	  replayExpirationDate: replay_expiration_date
	  rounds(order_by: {id: asc}) {
		id
		winReason: win_reason
		startedAt: started_at
		finishedAt: finished_at
		matchMapId: match_map_id
		spawnedPlayers: spawned_players
		winMatchTeamId: win_match_team_id
		__typename
	  }
	  maps(order_by: {number: asc}) {
		...MatchMapPrimaryParts
		replays {
		  ...MatchMapReplayPrimaryParts
		  __typename
		}
		map {
		  ...MapPrimaryParts
		  ...MapSecondaryParts
		  __typename
		}
		__typename
	  }
	  gameMode {
		id
		teamSize: team_size
		firstTeamSize: first_team_size
		secondTeamSize: second_team_size
		__typename
	  }
	  teams(order_by: {id: asc}) {
		...MatchTeamPrimaryParts
		mapStats {
		  ...MatchTeamMapStatPrimaryParts
		  __typename
		}
		__typename
	  }
	  members(order_by: {private: {party_id: desc_nulls_last}}) {
		...MatchMemberPrimaryParts
		private {
		  ...MatchMemberPrivateParts
		  __typename
		}
		__typename
	  }
	  __typename
	}
  }
  
  fragment MatchMapPrimaryParts on match_maps {
	id
	number
	mapId: map_id
	startedAt: started_at
	finishedAt: finished_at
	gameStatus: game_status
	__typename
  }
  
  fragment MatchMapReplayPrimaryParts on match_replays {
	id
	url
	createdAt: created_at
	__typename
  }
  
  fragment MapPrimaryParts on maps {
	id
	name
	__typename
  }
  
  fragment MapSecondaryParts on maps {
	offset
	scale
	preview
	topview
	overview
	flipV: flip_v
	flipH: flip_h
	__typename
  }
  
  fragment MatchTeamPrimaryParts on match_teams {
	id
	name
	size
	score
	chatId: chat_id
	isWinner: is_winner
	captainId: captain_id
	isDisqualified: is_disqualified
	__typename
  }
  
  fragment MatchTeamMapStatPrimaryParts on match_team_map_stats {
	score
	isWinner: is_winner
	matchMapId: match_map_id
	matchTeamId: match_team_id
	initialSide: initial_side
	__typename
  }
  
  fragment MatchMemberPrimaryParts on match_members {
	hash
	role
	ready
	impact
	connected
	isLeaver: is_leaver
	ratingDiff: rating_diff
	matchTeamId: match_team_id
	__typename
  }
  
  fragment MatchMemberPrivateParts on match_members_private {
	rating
	partyId: party_id
	user {
	  ...UserPrimaryParts
	  ...UserMediaParts
	  ...BasicUserSubscriptionParts
	  ...UserGameStats
	  __typename
	}
	__typename
  }
  
  fragment UserPrimaryParts on users {
	id
	link
	avatar
	online
	verified
	isMobile: is_mobile
	nickName: nick_name
	animatedAvatar: animated_avatar
	__typename
  }
  
  fragment UserMediaParts on users {
	isMedia: is_media
	displayMediaStatus: display_media_status
	__typename
  }
  
  fragment BasicUserSubscriptionParts on users {
	privacyOnlineStatusVisibility: privacy_online_status_visibility
	subscription {
	  planId: plan_id
	  __typename
	}
	icon {
	  ...ProfileIconPrimaryParts
	  __typename
	}
	__typename
  }
  
  fragment ProfileIconPrimaryParts on profile_icons {
	id
	url
	__typename
  }
  
  fragment UserGameStats on users {
	stats(
	  where: {game_id: {_eq: $gameId}, map_id: {_is_null: true}, game_mode_id: {_is_null: false}}
	) {
	  ...UserStatsParts
	  __typename
	}
	__typename
  }
  
  fragment UserStatsParts on user_stats {
	kills
	deaths
	place
	rating
	winRate: win_rate
	gameModeId: game_mode_id
	__typename
  }`
