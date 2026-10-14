export function redirectAfterLogin(next: string | null): string {
  return next || '/';
}
