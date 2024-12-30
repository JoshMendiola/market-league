import { Injectable } from '@angular/core';
import { environment } from '../../../environments/environment';
import { BehaviorSubject, Observable, map } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { League } from '../../models/league.model'; // Assuming you have this model defined
import { User } from '../../models/user.model';
import { Leagues } from '../../models/leagues.model';
import { Portfolio } from '../../models/portfolio.model';
import { Trade } from '../../models/trade.model';
import { devLog } from '../../../environments/development/devlog';
import { LeaguePortfolio } from '../../models/league-portfolio';

@Injectable({
  providedIn: 'root'
})
export class LeagueService {
  private baseUrl = environment.api_url;
  private findUserLeaguesUrl = `${this.baseUrl}/api/users/user-leagues`;
  private findLeagueMembersUrl = `${this.baseUrl}/api/leagues/details`;
  private createLeagueUrl = `${this.baseUrl}/api/leagues/create-league`;
  private removeLeagueUrl = `${this.baseUrl}/api/leagues/remove-league`;
  
  // BehaviorSubject for managing the selected league
  private selectedLeagueSource = new BehaviorSubject<League | null>(this.getStoredLeague());
  selectedLeague$ = this.selectedLeagueSource.asObservable();

  constructor(private http: HttpClient) {}

  // Method to get the user's leagues using their user ID
  getUserLeagues(userId: number): Observable<Leagues> {
    return this.http.post<Leagues>(this.findUserLeaguesUrl, { user_id: userId });
  }

  // Method to get members of a league using the league ID
  getLeagueMembers(leagueId: number): Observable<User[]> {
    return this.http.post<League>(this.findLeagueMembersUrl, { league_id: leagueId }).pipe(
      map((league: League) => league.users || []) // Ensure 'users' is not null or undefined, return empty array if it is
    );
  }

  // Method to set the selected league
  setSelectedLeague(league: League | null): void {
    devLog("Selected League: ", league);
    this.selectedLeagueSource.next(league); // Set the selected league as the full League object
    if (league) {
      // Store the entire league object as a JSON string in localStorage
      localStorage.setItem('selectedLeague', JSON.stringify(league)); 
    } else {
      localStorage.removeItem('selectedLeague');
    }
  }

  // Retrieve the stored league from localStorage (if it exists)
  getStoredLeague(): League | null {
    const storedLeague = localStorage.getItem('selectedLeague');
    
    // Check if storedLeague is a valid JSON
    if (storedLeague) {
      try {
        return JSON.parse(storedLeague) as League;
      } catch (e) {
        console.error("Error parsing stored league JSON:", e);
        localStorage.removeItem('selectedLeague'); // Clean up invalid entry
        return null;
      }
    }

    return null;
  }

  // Create League
  createLeague(leagueName: string, ownerUser: number, endDate: string): Observable<any> {
    const payload = {
      league_name: leagueName,
      owner_user: ownerUser,
      end_date: endDate
    };
    return this.http.post<any>(this.createLeagueUrl, payload); // Send POST request
  }

  // Remove League
  removeLeague(leagueId: number): Observable<any> {
    const payload = { league_id: leagueId }; // Payload with league_id
    return this.http.post<any>(this.removeLeagueUrl, payload); // Send POST request
  }

}
