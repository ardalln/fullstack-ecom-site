export const API = '/api/v1';
export const SESSION_KEY = 'shop.session';
export const CART_KEY = 'kicks.cart';

export function readSession() {
  try {
    const value = JSON.parse(localStorage.getItem(SESSION_KEY));
    if (value?.token && new Date(value.expiresAt) > new Date()) return value;
  } catch { /* empty storage */ }
  return null;
}

export function saveSession(value) {
  try {
    if (value) localStorage.setItem(SESSION_KEY, JSON.stringify(value));
    else localStorage.removeItem(SESSION_KEY);
  } catch { /* private browsing may disable storage */ }
}

export async function api(path, { method = 'GET', body, auth = true } = {}) {
  const headers = { Accept: 'application/json' };
  const token = auth ? readSession()?.token : null;
  if (token) headers.Authorization = `Bearer ${token}`;
  let payload = body;
  if (body !== undefined && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
    payload = JSON.stringify(body);
  }
  const response = await fetch(`${API}${path}`, { method, headers, body: payload });
  const data = response.status === 204 ? null : await response.json().catch(() => null);
  if (!response.ok) {
    const message = data?.message || data?.error || `خطای سرور (${response.status})`;
    const error = new Error(translateError(message));
    error.details = data?.details || null;
    error.status = response.status;
    throw error;
  }
  return data;
}

export function translateError(message) {
  const known = {
    'missing or invalid access token': 'برای ادامه وارد حساب کاربری شوید.',
    'you do not have permission to perform this action': 'این بخش فقط برای مدیر فروشگاه در دسترس است.',
    'resource not found': 'مورد درخواستی پیدا نشد.',
    'phone number is already registered': 'این شماره قبلاً ثبت شده است.',
    'username is already in use': 'این نام کاربری قبلاً انتخاب شده است.',
    'verification code is incorrect': 'کد تایید اشتباه است.',
    'verification code has expired': 'کد تایید منقضی شده؛ دوباره درخواست کنید.',
    'no verification code was requested for this phone number': 'ابتدا کد ورود را درخواست کنید.',
    'please wait before requesting another code': 'برای درخواست کد تازه کمی صبر کنید.',
    'rate limit exceeded': 'تعداد درخواست‌ها زیاد است؛ کمی بعد دوباره تلاش کن.',
    'insufficient stock': 'موجودی این محصول کافی نیست.',
    'coupon code is invalid or no longer available': 'کد تخفیف معتبر نیست یا مهلت استفاده از آن تمام شده است.',
    'order total does not meet the coupon minimum': 'مبلغ سبد خرید به حداقل لازم برای این کد تخفیف نرسیده است.',
    'shipping method is unavailable for this destination': 'این روش ارسال برای نشانی انتخاب‌شده در دسترس نیست.',
    'product slug is already in use': 'این نشانی محصول قبلاً استفاده شده است.',
    'product variant is used by a pending order and cannot be removed': 'این سایز در یک سفارش پرداخت‌نشده قرار دارد.',
    'ticket is closed': 'این تیکت بسته شده است.',
  };
  for (const [key, value] of Object.entries(known)) if (message.toLowerCase().includes(key)) return value;
  return message;
}

export const fa = (value) => new Intl.NumberFormat('fa-IR').format(Number(value || 0));
export const money = (value) => `${fa(value)} تومان`;
// Accept Persian/Arabic digits and common thousands separators in money fields.
// Return null for blank or malformed input so callers never turn it into zero.
export function parseWholeNumber(value) {
  const digits = String(value ?? '').replace(/[۰-۹٠-٩]/g, (digit) => {
    const code = digit.charCodeAt(0);
    return String(code >= 0x06f0 ? code - 0x06f0 : code - 0x0660);
  });
  const normalized = digits.replace(/[\s,٬،]/g, '');
  if (!/^\d+$/.test(normalized)) return null;
  const number = Number(normalized);
  return Number.isSafeInteger(number) ? number : null;
}
export const shortDate = (value) => value ? new Intl.DateTimeFormat('fa-IR', { dateStyle: 'medium' }).format(new Date(value)) : '—';
export const dateTime = (value) => value ? new Intl.DateTimeFormat('fa-IR', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) : '—';
export const productImage = (product) => product?.image_urls?.[0] || product?.image_url || '';

export function activePrice(product, variant = null) {
  const source = variant || product;
  return Number(source.discount_price) > 0 ? Number(source.discount_price) : Number(source.price || 0);
}

export function discountPercent(source) {
  const price = Number(source?.price || 0);
  const sale = Number(source?.discount_price || 0);
  if (!price || !sale || sale >= price) return 0;
  return Math.floor(((price - sale) * 100) / price);
}

export function parseRoute() {
  const hash = window.location.hash;
  const raw = hash.startsWith('#/') ? hash.slice(1)
    : /^\/blog(?:\/|$)/.test(window.location.pathname) ? window.location.pathname
    : /^\/(?:admin|shop|products|categories|brands|account|login|checkout|pay|track|contact|about|terms)(?:\/|$)/.test(window.location.pathname) ? window.location.pathname : '/';
  const [path, query = ''] = raw.split('?');
  const params = new URLSearchParams(query || (!hash ? window.location.search.slice(1) : ''));
  return { path: path || '/', params };
}

export function routeHref(path) {
  if (path === '/blog' || path.startsWith('/blog/')) return path;
  if (window.location.pathname !== '/') return `/#${path}`;
  return `#${path}`;
}

export function go(path) {
  if (path === '/blog' || path.startsWith('/blog/')) {
    window.history.pushState({}, '', path);
    window.dispatchEvent(new PopStateEvent('popstate'));
    return;
  }
  if (window.location.pathname !== '/') {
    window.history.pushState({}, '', `/#${path}`);
    window.dispatchEvent(new PopStateEvent('popstate'));
    return;
  }
  window.location.hash = path;
}

export function createSlug(value) {
  const persian = { ا: 'a', آ: 'a', ب: 'b', پ: 'p', ت: 't', ث: 's', ج: 'j', چ: 'ch', ح: 'h', خ: 'kh', د: 'd', ذ: 'z', ر: 'r', ز: 'z', ژ: 'zh', س: 's', ش: 'sh', ص: 's', ض: 'z', ط: 't', ظ: 'z', ع: 'a', غ: 'gh', ف: 'f', ق: 'gh', ک: 'k', گ: 'g', ل: 'l', م: 'm', ن: 'n', و: 'v', ه: 'h', ی: 'y' };
  return [...String(value || '').toLowerCase()].map((char) => persian[char] || char).join('').normalize('NFKD')
    .replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
}
