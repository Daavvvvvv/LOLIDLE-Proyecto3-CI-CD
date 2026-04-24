export interface ChampionListItem {
  id: string;
  name: string;
  imageKey: string;
}

export interface Champion {
  id: string;
  name: string;
  imageKey: string;
  gender: string;
  positions: string[];
  species: string;
  resource: string;
  rangeType: string;
  regions: string[];
  releaseYear: number;
}

export type AttributeStatus = 'match' | 'partial' | 'nomatch' | 'higher' | 'lower';

export interface AttributeFeedback {
  status: AttributeStatus;
}

export interface Feedback {
  gender: AttributeFeedback;
  positions: AttributeFeedback;
  species: AttributeFeedback;
  resource: AttributeFeedback;
  rangeType: AttributeFeedback;
  regions: AttributeFeedback;
  releaseYear: AttributeFeedback;
}

export interface CreateGameResponse {
  gameId: string;
}

export interface GuessResponse {
  guess: Champion;
  feedback: Feedback;
  correct: boolean;
  attemptCount: number;
}
