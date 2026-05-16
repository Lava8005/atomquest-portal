// frontend/src/auth.ts (HACKATHON DEMO MODE)

export async function initializeAuth(): Promise<void> {
  // Simulating Microsoft library initialization
  return new Promise(resolve => setTimeout(resolve, 300));
}

export async function login(): Promise<{ username: string }> {
  // Simulating popup login delay
  return new Promise(resolve => 
    setTimeout(() => resolve({ username: 'Arjun.Kumar@atomquest.com' }), 800)
  );
}

export async function acquireToken(): Promise<string> {
  // Returns a fake secure token that our API bridge will attach to requests
  return "demo_sso_jwt_token_987654321";
}

export async function getUserRole(): Promise<"Employee" | "Manager" | "Admin"> {
  // Defaults to Employee for the Phase 1 demo
  return "Employee"; 
}