import type { RouteProps } from '@/shared/lib/location';
import { AccountPage } from './AccountPage';
import { AccountsIndex } from './AccountsIndex';

/** `#/accounts` · `#/accounts/<id>` */
export function AccountsRoute({ segments }: RouteProps) {
  const [accountId] = segments;
  return accountId ? <AccountPage accountId={accountId} /> : <AccountsIndex />;
}
