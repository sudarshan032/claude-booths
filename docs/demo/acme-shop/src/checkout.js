const coupons = { SAVE10: { percent: 10, expires: '2099-01-01' }, OLD: { percent: 50, expires: '2020-01-01' } };

export function total(cart, code) {
  const sum = cart.reduce((s, item) => s + item.price * item.qty, 0);
  if (!code) return sum;
  const c = coupons[code];
  if (!c) throw new Error('unknown coupon');
  if (new Date(c.expires) < new Date()) throw new Error('expired');
  return Math.round(sum * (100 - c.percent)) / 100;
}
