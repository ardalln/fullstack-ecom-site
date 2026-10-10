import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  ArrowDownLeft, ArrowLeft, ArrowRight, ArrowUpLeft, BadgePercent, Check, ChevronDown,
  CircleHelp, Clock3, FileText, Gem, ImagePlus, LogOut, MapPin, Menu, Minus,
  Package, Plus, Printer, Search, Send, ShieldCheck, ShoppingBag, SlidersHorizontal, Sparkles,
  Trash2, Truck, UserRound, X,
} from 'lucide-react';
import {
  api, activePrice, CART_KEY, createSlug, dateTime, discountPercent, fa, go,
  money, parseRoute, parseWholeNumber, productImage, readSession, routeHref, saveSession, shortDate,
} from './api.js';
import { IRAN_LOCATIONS, IRAN_PROVINCES } from './data/iranLocations.js';

const ORDER_LABELS = {
  awaiting_payment: 'در انتظار پرداخت', paid: 'پرداخت‌شده', processing: 'در حال آماده‌سازی',
  shipped: 'ارسال‌شده', cancelled: 'لغوشده',
};

const DEMO_PREVIEW_MODE = (() => {
  const flag = new URLSearchParams(window.location.search).get('demo');
  try {
    if (flag === '0') { window.sessionStorage.removeItem('maleki.demo-preview'); return false; }
    if (flag === '1') window.sessionStorage.setItem('maleki.demo-preview', '1');
    return flag === '1' || window.sessionStorage.getItem('maleki.demo-preview') === '1';
  } catch { return flag === '1'; }
})();

const DEMO_CATEGORIES = [
  { id: 'demo-rings', name: 'انگشتر', slug: 'rings', parent_id: null, show_on_home: true },
  { id: 'demo-necklaces', name: 'گردنبند', slug: 'necklaces', parent_id: null, show_on_home: true },
  { id: 'demo-earrings', name: 'گوشواره', slug: 'earrings', parent_id: null, show_on_home: true },
  { id: 'demo-bracelets', name: 'دستبند', slug: 'bracelets', parent_id: null, show_on_home: true },
];

const DEMO_BRANDS = [{ id: 'demo-maleki', name: 'ملکی', slug: 'maleki' }];

const DEMO_PRODUCTS = [
  {
    id: 'demo-ring', name: 'انگشتر طلای آرکا', slug: 'demo-gold-ring',
    description: '<p>انگشتر آرکا با فرم ساده و پرداخت براق، برای استفادهٔ روزمره و هدیه انتخابی ماندگار است.</p><h3>ویژگی‌های محصول</h3><ul><li>طلای ۱۸ عیار</li><li>وزن تقریبی: ۲٫۳ گرم</li><li>ساخت گالری ملکی</li></ul>',
    price: 29800000, discount_price: 26900000, stock: 7, image_url: '/images/demo-ring.svg',
    image_urls: ['/images/demo-ring.svg', '/images/demo-ring-detail.svg'],
    category_id: 'demo-rings', category_name: 'انگشتر', category_slug: 'rings',
    brand_id: 'demo-maleki', brand_name: 'ملکی', brand_slug: 'maleki',
    variant_label: 'سایز انگشتر',
    variants: [
      { id: 'demo-ring-54', name: '۵۴', price: 29800000, discount_price: 26900000, stock: 4 },
      { id: 'demo-ring-56', name: '۵۶', price: 31200000, discount_price: 27900000, stock: 3, image_url: '/images/demo-ring-detail.svg' },
    ],
    attributes: { عیار: '۱۸ عیار', 'وزن تقریبی': '۲٫۳ گرم', 'مناسب برای': 'استفادهٔ روزمره و هدیه' },
    demo_only: true, is_popular: true,
  },
  {
    id: 'demo-necklace', name: 'گردنبند طلای ماه‌تاب', slug: 'demo-moonlight-necklace',
    description: '<p>گردنبند ماه‌تاب با زنجیر ظریف و آویز سنگی، در کنار استایل روزانه جلوه‌ای لطیف دارد.</p><h3>ویژگی‌های محصول</h3><ul><li>طلای ۱۸ عیار</li><li>طول زنجیر: ۴۰ سانتی‌متر</li><li>ساخت گالری ملکی</li></ul>',
    price: 36500000, discount_price: 32900000, stock: 5, image_url: '/images/demo-necklace.svg',
    image_urls: ['/images/demo-necklace.svg', '/images/demo-necklace-detail.svg'],
    category_id: 'demo-necklaces', category_name: 'گردنبند', category_slug: 'necklaces',
    brand_id: 'demo-maleki', brand_name: 'ملکی', brand_slug: 'maleki',
    variant_label: 'طول زنجیر',
    variants: [
      { id: 'demo-necklace-40', name: '۴۰ سانتی‌متر', price: 36500000, discount_price: 32900000, stock: 3 },
      { id: 'demo-necklace-45', name: '۴۵ سانتی‌متر', price: 38200000, discount_price: 34700000, stock: 2, image_url: '/images/demo-necklace-detail.svg' },
    ],
    attributes: { عیار: '۱۸ عیار', 'طول زنجیر': '۴۰ سانتی‌متر', 'نوع آویز': 'سنگی' },
    demo_only: true, is_popular: true,
  },
];

function DemoPreviewNotice() {
  if (!DEMO_PREVIEW_MODE) return null;
  return <div className="demo-preview-notice"><Sparkles size={15} /><span>پیش‌نمایش طراحی — این دو محصول نمونه‌اند و امکان خرید واقعی ندارند.</span></div>;
}

function readCart() {
  try {
    const cart = JSON.parse(localStorage.getItem(CART_KEY));
    if (Array.isArray(cart)) return cart;
    const legacy = JSON.parse(localStorage.getItem('shop.cart')) || {};
    return Object.values(legacy).map((line) => ({
      key: line.cartKey || `${line.id}|${line.variant_id || ''}`,
      product_id: Number(line.id || line.product_id), variant_id: line.variant_id || '', slug: line.slug || '',
      name: line.name || line.product_name || 'محصول', variant_name: line.variant_name || '',
      variant_label: line.variant_label || '', image_url: line.image_url || '', price: Number(line.price || 0),
      discount_price: 0, stock: Number(line.stock || 99), quantity: Number(line.qty || line.quantity || 1),
    }));
  } catch { return []; }
}

function normalizeBlogHash() {
  if (/^\/blog(?:\/|$)/.test(location.pathname) && location.hash.startsWith('#/')) {
    history.replaceState(history.state, '', `/${location.hash}`);
  }
}

function makeAnalyticsID() {
  if (window.crypto?.randomUUID) return window.crypto.randomUUID();
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (char) => {
    const random = Math.floor(Math.random() * 16);
    return (char === 'x' ? random : (random & 0x3) | 0x8).toString(16);
  });
}

function getAnalyticsVisitorID() {
  const key = 'outside.analytics.visitor';
  try {
    let value = localStorage.getItem(key);
    if (!value || !/^[0-9a-f-]{36}$/i.test(value)) {
      value = makeAnalyticsID();
      localStorage.setItem(key, value);
    }
    return value;
  } catch { return makeAnalyticsID(); }
}

function analyticsPagePath(path) {
  return path.replace(/^\/(?:pay|track)\/[^/]+/, (prefix) => prefix.startsWith('/pay/') ? '/pay' : '/track')
    .replace(/\/(?:orders|tickets)\/\d+(?:\/.*)?$/, (match) => match.replace(/\/\d+/, '/:id'));
}

export default function App() {
  const [route, setRoute] = useState(parseRoute);
  const [session, setSession] = useState(readSession);
  const [cart, setCart] = useState(readCart);
  const [categories, setCategories] = useState([]);
  const [brands, setBrands] = useState([]);
  const [searchValue, setSearchValue] = useState(route.params.get('q') || '');
  const [toasts, setToasts] = useState([]);

  const refreshRoute = useCallback(() => {
    normalizeBlogHash();
    const next = parseRoute();
    setRoute(next);
    setSearchValue(next.params.get('q') || '');
  }, []);

  useEffect(() => {
    normalizeBlogHash();
    window.addEventListener('hashchange', refreshRoute);
    window.addEventListener('popstate', refreshRoute);
    const refreshCatalog = (event) => {
      if (DEMO_PREVIEW_MODE) { setCategories(DEMO_CATEGORIES); setBrands(DEMO_BRANDS); return; }
      if (event.detail?.categories) setCategories(event.detail.categories);
      else api('/categories', { auth: false }).then(setCategories).catch(() => {});
      if (event.detail?.brands) setBrands(event.detail.brands);
      else api('/brands', { auth: false }).then(setBrands).catch(() => {});
    };
    refreshCatalog({ detail: null });
    window.addEventListener('catalog-refresh', refreshCatalog);
    return () => {
      window.removeEventListener('hashchange', refreshRoute);
      window.removeEventListener('popstate', refreshRoute);
      window.removeEventListener('catalog-refresh', refreshCatalog);
    };
  }, [refreshRoute]);

  const lastAnalyticsPath = useRef('');
  useEffect(() => {
    const visitorID = getAnalyticsVisitorID();
    const path = analyticsPagePath(route.path);
    const heartbeat = () => api('/analytics/heartbeat', { method: 'POST', auth: false, body: { visitor_id: visitorID, path } }).catch(() => {});
    heartbeat();
    const timer = window.setInterval(heartbeat, 25000);
    if (lastAnalyticsPath.current !== path) {
      lastAnalyticsPath.current = path;
      api('/analytics/view', { method: 'POST', auth: false, body: { visitor_id: visitorID, view_id: makeAnalyticsID(), path } }).catch(() => {});
    }
    return () => window.clearInterval(timer);
  }, [route.path]);

  useEffect(() => {
    try { localStorage.setItem(CART_KEY, JSON.stringify(cart)); } catch { /* storage may be unavailable */ }
  }, [cart]);

  useEffect(() => {
    const path = route.path;
    if (path === '/blog' || path.startsWith('/blog/')) return;
    document.title = path.startsWith('/admin') ? 'مدیریت فروشگاه | Maleki' : `${routeTitle(path)} | Maleki Jewelry Gallery`;
  }, [route.path]);

  const toast = useCallback((message, type = 'success') => {
    const id = `${Date.now()}-${Math.random()}`;
    setToasts((items) => [...items, { id, message, type }]);
    window.setTimeout(() => setToasts((items) => items.filter((item) => item.id !== id)), 3600);
  }, []);

  const updateSession = useCallback((value) => {
    saveSession(value);
    setSession(value);
  }, []);

  const addToCart = useCallback((product, variant = null) => {
    const key = `${product.id}|${variant?.id || ''}`;
    const source = variant || product;
    setCart((items) => {
      const existing = items.find((item) => item.key === key);
      if (existing) return items.map((item) => item.key === key
        ? { ...item, quantity: Math.min(item.quantity + 1, Math.max(1, item.stock)) } : item);
      return [...items, {
        key, product_id: Number(product.id), variant_id: variant?.id || '', slug: product.slug,
        name: product.name, variant_name: variant?.name || '', variant_label: product.variant_label || '',
        image_url: variant?.image_url || productImage(product), price: Number(source.price || product.price || 0),
        discount_price: Number(source.discount_price || 0), stock: Number(source.stock ?? product.stock ?? 0), quantity: 1,
      }];
    });
    toast('به سبد خرید اضافه شد.');
  }, [toast]);

  const setQuantity = (key, quantity) => setCart((items) => items.map((item) => item.key === key
    ? { ...item, quantity: Math.max(1, Math.min(item.stock || 1, Number(quantity) || 1)) } : item));
  const removeFromCart = (key) => setCart((items) => items.filter((item) => item.key !== key));
  const navigate = useCallback((path) => go(path), []);
  const count = cart.reduce((sum, item) => sum + item.quantity, 0);

  const renderPage = () => {
    if (route.path === '/login') return <LoginPage onLogin={(value) => { updateSession(value); toast('خوش آمدید!'); navigate('/account'); }} />;
    if (route.path === '/checkout') return <CheckoutPage items={cart} session={session} setQuantity={setQuantity} remove={removeFromCart}
      onClear={() => setCart([])} onNavigate={navigate} onLogin={(value) => { updateSession(value); toast('خوش آمدید!'); navigate('/checkout'); }} toast={toast} />;
    if (route.path === '/account' || route.path.startsWith('/account/')) return session
      ? <AccountPage route={route} session={session} toast={toast} onLogout={() => { updateSession(null); navigate('/'); }} />
      : <LoginPage onLogin={(value) => { updateSession(value); toast('خوش آمدید!'); navigate('/account'); }} />;
    if (route.path === '/admin' || route.path.startsWith('/admin/')) return session?.user?.role === 'admin'
      ? <AdminPage route={route} categories={categories} brands={brands} toast={toast} />
      : <LoginPage onLogin={(value) => { updateSession(value); navigate('/admin'); }} admin />;
    if (route.path.startsWith('/products/')) return <ProductPage slug={route.path.split('/')[2]} session={session} onAdd={addToCart} toast={toast} onNavigate={navigate} />;
    if (route.path.startsWith('/blog')) return <BlogPage path={route.path} onNavigate={navigate} />;
    if (route.path.startsWith('/pay/')) return <PaymentPage authority={route.path.split('/')[2]} toast={toast} onNavigate={navigate} />;
    if (route.path.startsWith('/track/')) return <TrackingPage code={route.path.split('/')[2]} />;
    if (['/contact', '/about', '/terms'].includes(route.path)) return <SiteContentPage page={route.path.slice(1)} session={session} toast={toast} onNavigate={navigate} />;
    if (route.path === '/' && !route.params.get('q')) return <HomePage categories={categories} brands={brands} onAdd={addToCart} />;
    if (route.path === '/shop' || route.path.startsWith('/categories/') || route.path.startsWith('/brands/') || route.path === '/') {
      return <StorePage route={route} categories={categories} brands={brands} onAdd={addToCart} />;
    }
    return <NotFound onNavigate={navigate} />;
  };

  const showMobileQuickNav = !route.path.startsWith('/admin');
  return <div className={`site-shell${showMobileQuickNav ? ' has-mobile-quick-nav' : ''}`}>
    <div className="announcement"><span>گالری طلا و زیورآلات ملکی</span><span className="announcement-side">زیبایی در جزئیات ماندگار است</span></div>
    <Header route={route} session={session} cartCount={count} categories={categories} brands={brands}
      searchValue={searchValue} setSearchValue={setSearchValue} onSearch={(q) => navigate(`/shop${q ? `?q=${encodeURIComponent(q)}` : ''}`)}
      onCart={() => navigate('/checkout')} onLogout={() => { updateSession(null); toast('از حساب خارج شدید.', 'info'); navigate('/'); }} />
    <main className="main-content" key={`${route.path}?${route.params.toString()}`}>{renderPage()}</main>
    <Footer onNavigate={navigate} />
    {showMobileQuickNav && <MobileQuickNav route={route} session={session} categories={categories} brands={brands} />}
    <div className="toast-stack" aria-live="polite">{toasts.map((item) => <div key={item.id} className={`toast toast-${item.type}`}><Check size={17} />{item.message}</div>)}</div>
  </div>;
}

function routeTitle(path) {
  if (path.startsWith('/categories/')) return 'دسته‌بندی زیورآلات';
  if (path.startsWith('/brands/')) return 'برندهای گالری';
  return ({ '/': 'خانه', '/shop': 'فروشگاه طلا و زیورآلات', '/login': 'ورود و ثبت‌نام', '/checkout': 'تسویه‌حساب', '/account': 'حساب کاربری', '/contact': 'تماس با ما', '/about': 'درباره ما', '/terms': 'قوانین و مقررات' })[path] || 'فروشگاه طلا و زیورآلات';
}

function Header({ route, session, cartCount, categories, brands, searchValue, setSearchValue, onSearch, onCart, onLogout }) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [mobileSubmenu, setMobileSubmenu] = useState('');
  const roots = categories.filter((item) => !item.parent_id);
  const closeMobileNavigation = () => { setMenuOpen(false); setMobileSubmenu(''); };
  const toggleMobileSubmenu = (name) => setMobileSubmenu((current) => current === name ? '' : name);
  const searchForm = <form className="header-search" onSubmit={(event) => { event.preventDefault(); onSearch(searchValue.trim()); }}>
    <Search size={18} aria-hidden="true" /><input value={searchValue} onChange={(event) => setSearchValue(event.target.value)} placeholder="جستجوی طلا، زیورآلات یا برند…" aria-label="جستجوی محصولات" />
  </form>;
  return <header className="header-wrap">
    <div className="header-main page-width">
      <a className="brand-lockup maleki-lockup" href={routeHref('/')} aria-label="Maleki Jewelry Gallery، صفحهٔ اصلی">
        <img className="maleki-logo" src="/images/maleki-header-hd.png" alt="Maleki Jewelry Gallery" />
      </a>
      <nav className={`primary-nav ${menuOpen ? 'is-open' : ''}`} aria-label="منوی اصلی">
        <a className={route.path === '/' ? 'nav-active' : ''} href={routeHref('/')} onClick={closeMobileNavigation}>خانه</a>
        <a className={route.path === '/shop' || route.path.startsWith('/categories/') || route.path.startsWith('/brands/') ? 'nav-active' : ''} href={routeHref('/shop')} onClick={closeMobileNavigation}>فروشگاه</a>
        <div className={`nav-menu ${mobileSubmenu === 'categories' ? 'menu-open' : ''}`}>
          <a className="desktop-nav-menu-link" href={routeHref('/shop')} onClick={closeMobileNavigation}>دسته‌بندی‌ها <ChevronDown size={14} /></a>
          <button className="mobile-nav-menu-toggle" type="button" aria-expanded={mobileSubmenu === 'categories'} onClick={() => toggleMobileSubmenu('categories')}>دسته‌بندی‌ها <ChevronDown size={14} /></button>
          <div className="nav-menu-popover">
            {roots.map((root) => <div className="nav-menu-group" key={root.id}>
              <a className="nav-menu-root" href={routeHref(`/categories/${encodeURIComponent(root.slug)}`)} onClick={closeMobileNavigation}>{root.name}</a>
              {categories.filter((item) => String(item.parent_id) === String(root.id)).map((child) => <a key={child.id} href={routeHref(`/categories/${encodeURIComponent(child.slug)}`)} onClick={closeMobileNavigation}>{child.name}</a>)}
            </div>)}
            {!roots.length && <span className="muted">هنوز دسته‌ای ثبت نشده است.</span>}
          </div>
        </div>
        <div className={`nav-menu ${mobileSubmenu === 'brands' ? 'menu-open' : ''}`}>
          <a className="desktop-nav-menu-link" href={routeHref('/shop')} onClick={closeMobileNavigation}>برندها <ChevronDown size={14} /></a>
          <button className="mobile-nav-menu-toggle" type="button" aria-expanded={mobileSubmenu === 'brands'} onClick={() => toggleMobileSubmenu('brands')}>برندها <ChevronDown size={14} /></button>
          <div className="nav-menu-popover nav-brand-popover">{brands.map((brand) => <a key={brand.id} href={routeHref(`/brands/${encodeURIComponent(brand.slug)}`)} onClick={closeMobileNavigation}>{brand.name}</a>)}{!brands.length && <span className="muted">هنوز برندی ثبت نشده است.</span>}</div>
        </div>
        <a href="/blog" onClick={closeMobileNavigation}>مجله</a>
      </nav>
      <div className="header-actions">
        <button className="icon-button mobile-menu-button" type="button" aria-label={menuOpen ? 'بستن فهرست' : 'باز کردن فهرست'} aria-expanded={menuOpen} onClick={() => { setMenuOpen((value) => !value); setMobileSubmenu(''); }}>{menuOpen ? <X size={19} /> : <Menu size={20} />}<span>{menuOpen ? 'بستن' : 'منو'}</span></button>
        {session ? <><a className="account-link" href={routeHref('/account')}><UserRound size={18} /><span>{session.user?.first_name || 'حساب من'}</span></a><button className="icon-button logout-icon" onClick={onLogout} title="خروج" aria-label="خروج"><LogOut size={18} /></button></>
          : <a className="account-link" href={routeHref('/login')}><UserRound size={18} /><span>ورود</span></a>}
        <button className="header-cart" type="button" onClick={onCart} aria-label={`سبد خرید، ${cartCount} کالا`}><ShoppingBag size={19} /><span>سبد خرید</span><b>{fa(cartCount)}</b></button>
      </div>
      {searchForm}
    </div>
  </header>;
}

function MobileQuickNav({ route, session, categories, brands }) {
  const [selectedPanel, setSelectedPanel] = useState('categories');
  const [panelOpen, setPanelOpen] = useState(false);
  const roots = categories.filter((item) => !item.parent_id);

  useEffect(() => { setPanelOpen(false); }, [route.path]);
  useEffect(() => {
    if (!panelOpen) return undefined;
    const closeOnEscape = (event) => { if (event.key === 'Escape') setPanelOpen(false); };
    window.addEventListener('keydown', closeOnEscape);
    return () => window.removeEventListener('keydown', closeOnEscape);
  }, [panelOpen]);

  const panelLabel = selectedPanel === 'categories' ? 'دسته‌بندی‌ها' : 'برندها';
  const togglePanel = (panel) => {
    if (panelOpen && selectedPanel === panel) setPanelOpen(false);
    else { setSelectedPanel(panel); setPanelOpen(true); }
  };
  return <div className="mobile-quick-nav">
    <section className={`mobile-quick-panel${panelOpen ? ' is-open' : ''}`} id="mobile-quick-panel" aria-label={panelLabel} aria-hidden={!panelOpen} inert={!panelOpen}>
      <div className="mobile-quick-panel-heading"><strong>{panelLabel}</strong><button type="button" className="mobile-quick-close" onClick={() => setPanelOpen(false)} aria-label="بستن"><X size={18} /></button></div>
      {selectedPanel === 'categories' ? <div className="mobile-quick-category-list">
        <a className="mobile-quick-all" href={routeHref('/shop')}>مشاهدهٔ همهٔ محصولات <ArrowLeft size={15} /></a>
        {roots.map((root) => {
          const children = categories.filter((item) => String(item.parent_id) === String(root.id));
          return <div className="mobile-quick-category-group" key={root.id}>
            <a className="mobile-quick-category-root" href={routeHref(`/categories/${encodeURIComponent(root.slug)}`)}>{root.name}<ArrowLeft size={14} /></a>
            {children.length > 0 && <div className="mobile-quick-category-children">{children.map((child) => <a key={child.id} href={routeHref(`/categories/${encodeURIComponent(child.slug)}`)}>{child.name}</a>)}</div>}
          </div>;
        })}
        {!roots.length && <p className="muted">هنوز دسته‌بندی‌ای ثبت نشده است.</p>}
      </div> : <div className="mobile-quick-brand-list">
        {brands.map((brand) => <a key={brand.id} href={routeHref(`/brands/${encodeURIComponent(brand.slug)}`)}><Sparkles size={15} />{brand.name}</a>)}
        {!brands.length && <p className="muted">هنوز برندی ثبت نشده است.</p>}
      </div>}
    </section>
    <nav className="mobile-quick-bar" aria-label="دسترسی سریع">
      <button type="button" className={panelOpen && selectedPanel === 'categories' ? 'mobile-quick-active' : ''} aria-expanded={panelOpen && selectedPanel === 'categories'} aria-controls="mobile-quick-panel" onClick={() => togglePanel('categories')}><Gem size={19} /><span>دسته‌ها</span></button>
      <button type="button" className={panelOpen && selectedPanel === 'brands' ? 'mobile-quick-active' : ''} aria-expanded={panelOpen && selectedPanel === 'brands'} aria-controls="mobile-quick-panel" onClick={() => togglePanel('brands')}><Sparkles size={19} /><span>برندها</span></button>
      <a className={route.path.startsWith('/account') || route.path === '/login' ? 'mobile-quick-active' : ''} href={routeHref(session ? '/account' : '/login')}><UserRound size={19} /><span>پروفایل</span></a>
    </nav>
  </div>;
}

function useHorizontalDrag() {
  const drag = useRef(null);
  const suppressClick = useRef(false);

  const onPointerDown = (event) => {
    if (event.pointerType !== 'mouse' || event.button !== 0) return;
    drag.current = { startX: event.clientX, startScrollLeft: event.currentTarget.scrollLeft, moved: false };
  };
  const onPointerMove = (event) => {
    if (!drag.current) return;
    const delta = event.clientX - drag.current.startX;
    if (!drag.current.moved && Math.abs(delta) < 5) return;
    if (!drag.current.moved) event.currentTarget.setPointerCapture(event.pointerId);
    drag.current.moved = true;
    event.currentTarget.classList.add('is-dragging');
    const rtl = getComputedStyle(event.currentTarget).direction === 'rtl';
    event.currentTarget.scrollLeft = drag.current.startScrollLeft + (rtl ? delta : -delta);
    event.preventDefault();
  };
  const finishDrag = (event) => {
    if (!drag.current) return;
    if (drag.current.moved) {
      suppressClick.current = true;
      event.currentTarget.classList.remove('is-dragging');
      window.setTimeout(() => { suppressClick.current = false; }, 0);
    }
    drag.current = null;
  };
  const onClickCapture = (event) => {
    if (!suppressClick.current) return;
    event.preventDefault();
    event.stopPropagation();
    suppressClick.current = false;
  };
  return {
    onPointerDown, onPointerMove, onPointerUp: finishDrag,
    onPointerCancel: finishDrag, onClickCapture,
  };
}

function RotateCcwIcon() { return <ArrowDownLeft size={16} />; }

function Footer({ onNavigate }) {
  return <footer className="site-footer"><div className="page-width footer-main">
    <div className="footer-brand"><a className="brand-lockup maleki-lockup" href={routeHref('/')}><img className="maleki-logo" src="/images/maleki-header-hd.png" alt="Maleki Jewelry Gallery" /></a><p>زیورآلاتی برای لحظه‌های ماندگار؛ انتخابی ظریف برای هر روز.</p></div>
    <div><strong>راهنمای خرید</strong><a href={routeHref('/shop')}>همهٔ محصولات</a><a href={routeHref('/account/orders')}>پیگیری سفارش</a><a href={routeHref('/contact')}>تماس با ما و پشتیبانی</a></div>
    <div><strong>گالری ملکی</strong><a href="/blog">مجله و راهنما</a><a href={routeHref('/about')}>درباره ما</a><a href={routeHref('/terms')}>قوانین و مقررات</a><span className="muted">زیبایی در جزئیات است.</span></div>
    <div className="footer-newsletter"><strong>از ویترین تازه باخبر شو</strong><p>تازه‌های گالری و راهنمای انتخاب زیورآلات.</p><button className="text-link" onClick={() => onNavigate('/blog')}>رفتن به مجله <ArrowLeft size={15} /></button></div>
  </div><div className="page-width footer-bottom"><span>© Maleki Jewelry Gallery · ۱۴۰۵</span><span>زیبایی در جزئیات ماندگار می‌شود.</span></div></footer>;
}

function HomePage({ categories, brands, onAdd }) {
  const [products, setProducts] = useState([]);
  const [popularProducts, setPopularProducts] = useState([]);
  const [catalogError, setCatalogError] = useState('');
  const [popularError, setPopularError] = useState('');
  const [loading, setLoading] = useState(true);
  const latestDrag = useHorizontalDrag();
  const bestPriceDrag = useHorizontalDrag();
  const categoryDrag = useHorizontalDrag();
  useEffect(() => {
    let active = true;
    if (DEMO_PREVIEW_MODE) {
      setProducts(DEMO_PRODUCTS);
      setPopularProducts(DEMO_PRODUCTS);
      setCatalogError('');
      setPopularError('');
      setLoading(false);
      return () => { active = false; };
    }
    Promise.allSettled([
      api('/products?page=1&limit=12', { auth: false }),
      api('/products?popular=true&page=1&limit=12', { auth: false }),
    ])
      .then(([catalog, popular]) => {
        if (!active) return;
        if (catalog.status === 'fulfilled') {
          setProducts(catalog.value.data || []);
          setCatalogError('');
        } else {
          setCatalogError(catalog.reason?.message || 'دریافت فهرست محصولات ممکن نشد.');
        }
        if (popular.status === 'fulfilled') {
          setPopularProducts(popular.value.data || []);
          setPopularError('');
        } else {
          setPopularError(popular.reason?.message || 'دریافت محصولات محبوب ممکن نشد.');
        }
      })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);
  const available = products.filter((product) => product.stock > 0);
  const latest = popularProducts.filter((product) => product.stock > 0).slice(0, 12);
  const bestPrice = [...available].sort((a, b) => lowestPriceForSort(a) - lowestPriceForSort(b)).slice(0, 5);
  const homeCategories = categories.filter((item) => !item.parent_id && item.show_on_home).slice(0, 6);
  const defaultCategories = [
    { id: 'rings', name: 'انگشتر', slug: '' }, { id: 'necklaces', name: 'گردنبند', slug: '' },
    { id: 'earrings', name: 'گوشواره', slug: '' }, { id: 'bracelets', name: 'دستبند', slug: '' },
  ];
  const categoryItems = homeCategories.length ? homeCategories : defaultCategories;
  return <div className="maleki-home">
    <CampaignSlider />
    <section className="page-width jewelry-category-section">
      <div className="home-section-head"><div><span className="eyebrow">گالری ملکی</span><h2>دسته‌بندی‌های اصلی</h2></div><a className="text-link" href={routeHref('/shop')}>مشاهده همه <ArrowLeft size={15} /></a></div>
      <DemoPreviewNotice />
      <div className="jewelry-category-rail drag-scroll" aria-label="دسته‌بندی‌های منتخب" {...categoryDrag}>{categoryItems.map((item, index) => <a className="jewelry-category-tile" href={item.slug ? routeHref(`/categories/${encodeURIComponent(item.slug)}`) : routeHref('/shop')} key={item.id}>
        <span className="jewelry-category-visual">{item.home_image_url ? <img src={item.home_image_url} alt="" loading="lazy" /> : <Gem size={index % 2 ? 30 : 34} strokeWidth={1.2} />}</span>
        <b>{item.home_title || item.name}</b>
      </a>)}</div>
    </section>
    <section id="catalog" className="page-width home-edit-picks jewelry-product-section">
      <div className="home-section-head"><div><span className="eyebrow">تازه‌های ویترین</span><h2>انتخابی برای هر روز</h2><p>زیورآلاتی که با سلیقهٔ تو کامل می‌شوند.</p></div><a className="text-link" href={routeHref('/shop')}>رفتن به فروشگاه <ArrowLeft size={15} /></a></div>
      {loading ? <ProductSkeletons /> : popularError ? <ErrorPanel message={popularError} /> : latest.length ? <div className="home-product-rail drag-scroll" aria-label="محصولات محبوب ویترین" {...latestDrag}>{latest.map((product, index) => <ProductCard key={product.id} product={product} index={index} onAdd={onAdd} />)}</div> : <EmptyState icon={Gem} title="محصول محبوبی در ویترین نیست" body="برای نمایش در این بخش، گزینهٔ «محبوب» را در ویرایش محصول فعال کن." action={<a className="button button-outline" href={routeHref('/shop')}>رفتن به فروشگاه</a>} />}
    </section>
    <section className="page-width jewelry-secondary-banner">
      <div className="jewelry-secondary-copy"><span className="eyebrow">MALEKI · JEWELRY GALLERY</span><h2>زیبایی، در جزئیات ماندگار است.</h2><p>قطعه‌ای را انتخاب کن که روایتگر سلیقهٔ تو باشد.</p><a className="button button-green" href={routeHref('/shop')}>دیدن مجموعه <ArrowLeft size={16} /></a></div>
      <div className="jewelry-secondary-art" aria-hidden="true"><span className="jewelry-orbit jewelry-orbit-one" /><span className="jewelry-orbit jewelry-orbit-two" /><Gem size={106} strokeWidth={.8} /></div>
    </section>
    <section className="page-width home-popular jewelry-best-price">
      <div className="home-section-head"><div><span className="eyebrow">انتخاب هوشمندانه</span><h2>با بهترین قیمت</h2><p>زیبایی دلنشین، با انتخابی متناسب با بودجه.</p></div><a className="text-link" href={routeHref('/shop')}>دیدن همه <ArrowLeft size={15} /></a></div>
      {loading ? <ProductSkeletons /> : catalogError ? <ErrorPanel message={catalogError} /> : bestPrice.length ? <div className="best-price-rail drag-scroll" aria-label="محصولات با بهترین قیمت" {...bestPriceDrag}>{bestPrice.map((product, index) => <ProductCard key={product.id} product={product} index={index} onAdd={onAdd} />)}</div> : <p className="muted">محصولی برای نمایش موجود نیست.</p>}
    </section>
    {brands.length > 0 && <section className="page-width home-brand-strip"><div><span className="eyebrow">نام‌های آشنا</span><h2>برندها و سازندگان</h2></div><div className="brand-links">{brands.slice(0, 8).map((brand) => <a key={brand.id} href={routeHref(`/brands/${encodeURIComponent(brand.slug)}`)}>{brand.name}<ArrowUpLeft size={14} /></a>)}</div></section>}
  </div>;
}

function StorePage({ route, categories, brands, onAdd }) {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [categoryName, setCategoryName] = useState('');
  const [brandName, setBrandName] = useState('');
  const [categoryData, setCategoryData] = useState(null);
  const [brandData, setBrandData] = useState(null);
  const [sort, setSort] = useState('newest');
  const [brandFilter, setBrandFilter] = useState('');
  const [sizeFilter, setSizeFilter] = useState('');
  const [filtersOpen, setFiltersOpen] = useState(false);
  const query = route.params.get('q') || '';
  const pathParts = route.path.split('/');
  const categorySlug = route.path.startsWith('/categories/') ? decodeURIComponent(pathParts[2] || '') : '';
  const brandSlug = route.path.startsWith('/brands/') ? decodeURIComponent(pathParts[2] || '') : '';

  useEffect(() => {
    let active = true;
    async function load() {
      setLoading(true); setError('');
      if (DEMO_PREVIEW_MODE) {
        const selectedCategory = DEMO_CATEGORIES.find((item) => item.slug === categorySlug) || null;
        const selectedBrand = DEMO_BRANDS.find((item) => item.slug === brandSlug) || null;
        const search = query.trim().toLocaleLowerCase('fa-IR');
        const previewProducts = DEMO_PRODUCTS.filter((product) =>
          (!categorySlug || product.category_slug === categorySlug) &&
          (!brandSlug || product.brand_slug === brandSlug) &&
          (!search || `${product.name} ${product.description} ${product.brand_name}`.toLocaleLowerCase('fa-IR').includes(search)));
        if (active) {
          setCategoryName(selectedCategory?.name || ''); setBrandName(selectedBrand?.name || '');
          setCategoryData(selectedCategory); setBrandData(selectedBrand); setProducts(previewProducts); setLoading(false);
        }
        return;
      }
      try {
        let categoryID = null; let brandID = null; let heading = '';
        let selectedCategory = null; let selectedBrand = null;
        if (categorySlug) {
          const category = categories.find((item) => item.slug === categorySlug) || await api(`/categories/slug/${encodeURIComponent(categorySlug)}`, { auth: false });
          categoryID = category.id; heading = category.name; selectedCategory = category;
        }
        if (brandSlug) {
          const brand = brands.find((item) => item.slug === brandSlug) || await api(`/brands/slug/${encodeURIComponent(brandSlug)}`, { auth: false });
          brandID = brand.id; heading = brand.name; selectedBrand = brand;
        }
        if (active) { setCategoryName(categorySlug ? heading : ''); setBrandName(brandSlug ? heading : ''); setCategoryData(selectedCategory); setBrandData(selectedBrand); }
        const params = new URLSearchParams({ page: '1', limit: '100' });
        if (query) params.set('q', query);
        if (categoryID) params.set('category_id', categoryID);
        if (brandID) params.set('brand_id', brandID);
        const result = await api(`/products?${params}`, { auth: false });
        if (active) setProducts(result.data || []);
      } catch (failure) { if (active) setError(failure.message); }
      finally { if (active) setLoading(false); }
    }
    load();
    return () => { active = false; };
  }, [categorySlug, brandSlug, query, categories, brands]);

  useEffect(() => {
    const item = brandData || categoryData;
    if (!item) return;
    document.title = `${item.seo_title || item.name} | ${brandData ? 'برند' : 'دسته‌بندی'} MALEKI`;
    let meta = document.querySelector('meta[name="description"]');
    if (!meta) { meta = document.createElement('meta'); meta.name = 'description'; document.head.append(meta); }
    meta.content = (item.seo_description || richTextPlainText(item.description)).slice(0, 320);
  }, [brandData, categoryData]);

  const sizeOptions = useMemo(() => {
    const values = new Map();
    for (const product of products) for (const value of productSizeValues(product)) {
      const key = normalizeSize(value);
      if (key && !values.has(key)) values.set(key, String(value));
    }
    return [...values.entries()].sort((a, b) => Number(a[0]) - Number(b[0]) || a[1].localeCompare(b[1])).map(([key, label]) => ({ key, label }));
  }, [products]);
  const matching = useMemo(() => products.filter((product) => {
    if (brandFilter && String(product.brand_id) !== brandFilter) return false;
    if (sizeFilter && !productSizeValues(product).some((value) => normalizeSize(value) === sizeFilter)) return false;
    return true;
  }), [products, brandFilter, sizeFilter]);
  const sorted = useMemo(() => [...matching].sort((a, b) => sort === 'price-low' ? lowestPriceForSort(a) - lowestPriceForSort(b) : sort === 'price-high' ? highestPriceForSort(b) - highestPriceForSort(a) : Number(b.id) - Number(a.id)), [matching, sort]);
  const title = categoryName || brandName || (query ? `نتایج جستجو برای «${query}»` : 'همهٔ زیورآلات');
  const hasFilters = Boolean(brandFilter || sizeFilter);
  const clearFilters = () => { setBrandFilter(''); setSizeFilter(''); };
  const activeCategory = categories.find((item) => item.slug === categorySlug);
  const activeRootCategory = activeCategory?.parent_id
    ? categories.find((item) => String(item.id) === String(activeCategory.parent_id))
    : activeCategory;
  const rootCategories = categories.filter((item) => !item.parent_id);
  const childCategories = activeRootCategory
    ? categories.filter((item) => String(item.parent_id) === String(activeRootCategory.id))
    : [];
  return <section id="catalog" className="page-width shop-section shop-page">
    <DemoPreviewNotice />
    <div className="catalog-heading"><div><span className="eyebrow"><Sparkles size={14} /> گالری طلا و زیورآلات</span><h1>{title}</h1><p>{query ? 'نتایج جست‌وجو را دقیق‌تر کن.' : 'از میان طلا و زیورآلات، انتخابت را پیدا کن.'}</p></div>
      <div className="catalog-heading-tools"><span className="catalog-result-count">{loading ? 'در حال دریافت محصولات…' : `${fa(sorted.length)} محصول`}</span><label className="sort-select"><SlidersHorizontal size={17} /><span>مرتب‌سازی</span><select aria-label="مرتب‌سازی محصولات" value={sort} onChange={(event) => setSort(event.target.value)}><option value="newest">جدیدترین</option><option value="price-low">ارزان‌ترین</option><option value="price-high">گران‌ترین</option></select></label></div>
    </div>
    {categories.length > 0 && <div className="category-chips" aria-label="دسته‌بندی‌ها">
      <div className="category-chips-desktop"><a className={!categorySlug ? 'chip chip-active' : 'chip'} href={routeHref('/shop')}>همهٔ مدل‌ها</a>{categories.map((item) => <a key={item.id} className={categorySlug === item.slug ? 'chip chip-active' : 'chip'} href={routeHref(`/categories/${encodeURIComponent(item.slug)}`)}>{item.name}</a>)}</div>
      <div className="category-chips-mobile">
        <div className="category-chip-parents"><a className={!categorySlug ? 'chip chip-active' : 'chip'} href={routeHref('/shop')}>همهٔ مدل‌ها</a>{rootCategories.map((item) => <a key={item.id} className={activeRootCategory?.id === item.id ? 'chip chip-active' : 'chip'} href={routeHref(`/categories/${encodeURIComponent(item.slug)}`)}>{item.name}</a>)}</div>
        {childCategories.length > 0 && <div className="category-chip-children" aria-label={`زیر دسته‌های ${activeRootCategory.name}`}>
          <a className={activeCategory?.id === activeRootCategory.id ? 'chip chip-active' : 'chip'} href={routeHref(`/categories/${encodeURIComponent(activeRootCategory.slug)}`)}>همهٔ {activeRootCategory.name}</a>
          {childCategories.map((item) => <a key={item.id} className={categorySlug === item.slug ? 'chip chip-active' : 'chip'} href={routeHref(`/categories/${encodeURIComponent(item.slug)}`)}>{item.name}</a>)}
        </div>}
      </div>
    </div>}
    <div className="catalog-layout">
      <aside className={`catalog-filters ${filtersOpen ? 'catalog-filters-open' : ''}`} aria-label="فیلتر محصولات">
        <div className="filter-panel-heading"><div><span className="eyebrow">جست‌وجوی دقیق</span><h2>فیلترها</h2></div><button type="button" aria-label="بستن فیلترها" className="icon-button filter-close" onClick={() => setFiltersOpen(false)}><X size={18} /></button></div>
        <section className="filter-group"><h3>برند</h3><div className="filter-brand-list">{brands.map((brand) => <label className="filter-option" key={brand.id}><input type="radio" name="store-brand" value={brand.id} checked={brandFilter === String(brand.id)} onChange={() => setBrandFilter(String(brand.id))} /><span>{brand.name}</span></label>)}</div>{!brands.length && <small className="filter-help">هنوز برندی ثبت نشده است.</small>}</section>
        <section className="filter-group"><div className="filter-group-heading"><h3>سایز</h3>{sizeFilter && <button type="button" className="filter-clear-one" onClick={() => setSizeFilter('')}>پاک‌کردن</button>}</div>
          {sizeOptions.length ? <div className="filter-size-list">{sizeOptions.map(({ key, label }) => <button type="button" key={key} className={`filter-size-chip ${sizeFilter === key ? 'filter-size-active' : ''}`} aria-pressed={sizeFilter === key} onClick={() => setSizeFilter((current) => current === key ? '' : key)}>{label}</button>)}</div> : <small className="filter-help">برای نمایش سایزها، محصولات باید تنوع سایز داشته باشند.</small>}
        </section>
        {hasFilters && <button type="button" className="button button-outline filter-reset" onClick={clearFilters}>حذف همهٔ فیلترها <X size={15} /></button>}
        <button type="button" className="button button-dark filter-apply" onClick={() => setFiltersOpen(false)}>نمایش {fa(sorted.length)} محصول</button>
      </aside>
      <div className="catalog-results">
        <div className="catalog-mobile-tools"><button className="button button-outline" type="button" onClick={() => setFiltersOpen(true)}><SlidersHorizontal size={16} /> فیلتر سایز و برند</button><span>{loading ? '…' : `${fa(sorted.length)} محصول`}</span></div>
        {error && <ErrorPanel message={error} />}
        {loading ? <ProductSkeletons /> : sorted.length ? <div className="product-grid">{sorted.map((product, index) => <ProductCard key={product.id} product={product} index={index} onAdd={onAdd} />)}</div>
          : <EmptyState icon={Search} title="مدلی پیدا نشد" body={hasFilters ? 'فیلتر برند یا سایز را پاک کن و دوباره ببین.' : 'عبارت جست‌وجو یا دسته‌بندی را تغییر بده.'} action={hasFilters ? <button className="button button-outline" type="button" onClick={clearFilters}>پاک‌کردن فیلترها</button> : <a className="button button-outline" href={routeHref('/shop')}>نمایش همهٔ محصولات</a>} />}
      </div>
    </div>
    {!loading && (brandData?.description || categoryData?.description) && <section className="catalog-seo-copy"><span className="eyebrow">{brandData ? `راهنمای برند ${brandData.name}` : `راهنمای دستهٔ ${categoryData.name}`}</span><h2>{brandData?.name || categoryData?.name}</h2><RichTextBlock value={brandData?.description || categoryData?.description} /></section>}
  </section>;
}

function normalizeSize(value) {
  return String(value ?? '').replace(/[۰-۹٠-٩]/g, (digit) => {
    const code = digit.charCodeAt(0);
    return String(code >= 0x06f0 ? code - 0x06f0 : code - 0x0660);
  }).trim().toLowerCase().replace(/^(سایز|اندازه|eu|شماره)\s*/i, '').replace(/\s+/g, ' ');
}

function productSizeValues(product) {
  const label = product.variant_label || '';
  const sizeLabel = /سایز|اندازه|شماره|\bsize\b|\beu\b|\bus\b/i.test(label);
  const otherAxis = /رنگ|حجم|ظرفیت|حافظه|وزن|جنس|\bcolor\b|\bvolume\b|\bcapacity\b/i.test(label);
  const values = [];
  for (const variant of product.variants || []) {
    const normalized = normalizeSize(variant.name);
    const numericSize = /^\d{1,2}(?:[.,]\d)?$/.test(normalized) && Number(normalized.replace(',', '.')) >= 4 && Number(normalized.replace(',', '.')) <= 55;
    if (sizeLabel || (!otherAxis && numericSize) || /^(eu|us)\s*\d/i.test(normalized)) values.push(variant.name);
  }
  for (const [key, value] of Object.entries(product.attributes || {})) {
    if (/سایز|اندازه|شماره|\bsize\b/i.test(key) && value) values.push(value);
  }
  return values;
}

function lowestPriceForSort(product) {
  const variants = product.variants || [];
  return variants.length ? Math.min(...variants.map(activePrice.bind(null, product))) : activePrice(product);
}

function highestPriceForSort(product) {
  const variants = product.variants || [];
  return variants.length ? Math.max(...variants.map(activePrice.bind(null, product))) : activePrice(product);
}

function TrustStrip() {
  return <div className="trust-strip"><div><ShieldCheck size={18} /><span><b>اصالت کالا</b><small>تضمین اصالت همه مدل‌ها</small></span></div><div><Truck size={18} /><span><b>ارسال سریع</b><small>تحویل به موقع، درب منزل</small></span></div><div><RotateCcwIcon /><span><b>بازگشت آسان</b><small>۷ روز برای تصمیم نهایی</small></span></div><div><CircleHelp size={18} /><span><b>پشتیبانی واقعی</b><small>هر وقت نیاز داشتی، کنارتیم</small></span></div></div>;
}

function ProductCard({ product, index = 0, onAdd }) {
  const variants = product.variants || [];
  const lowest = variants.length ? [...variants].sort((a, b) => activePrice(product, a) - activePrice(product, b))[0] : null;
  const source = lowest || product;
  const sale = Number(source.discount_price || 0);
  const percent = discountPercent(source);
  const finalPrice = activePrice(product, source === product ? null : source);
  const image = source.image_url || productImage(product);
  return <article className="product-card" style={{ '--card-delay': `${Math.min(index % 8, 7) * 45}ms` }}>
    <a className="product-card-image" href={routeHref(`/products/${encodeURIComponent(product.slug)}`)} aria-label={`مشاهده ${product.name}`}>
      {image ? <img className={product.demo_only ? 'demo-product-art' : ''} src={image} alt={product.name} loading="lazy" /> : <ProductPlaceholder />}
      {percent > 0 && <span className="discount-badge"><BadgePercent size={13} /> {fa(percent)}٪</span>}
      {product.demo_only && <span className="demo-product-badge">نمونهٔ نمایشی</span>}
      {product.stock <= 0 && <span className="out-badge">ناموجود</span>}
      <span className="card-image-cta">مشاهده مدل <ArrowUpLeft size={15} /></span>
    </a>
    <div className="product-card-content"><div className="product-card-meta"><a href={routeHref(`/brands/${encodeURIComponent(product.brand_slug || '')}`)}>{product.brand_name || 'ملکی'}</a><span>{product.category_name || 'زیورآلات'}</span></div>
      <h3><a href={routeHref(`/products/${encodeURIComponent(product.slug)}`)}>{product.name}</a></h3>
      <div className="card-bottom"><div className="price-stack">{sale > 0 ? <><del>{money(source.price)}</del><strong>{money(finalPrice)}</strong></> : <strong>{variants.length ? `از ${money(finalPrice)}` : money(finalPrice)}</strong>}</div>
        {product.demo_only ? <span className="demo-card-note">فقط پیش‌نمایش</span> : <button className="quick-add" type="button" disabled={!product.stock} aria-label={variants.length ? `انتخاب ${product.variant_label || 'گزینه'}` : 'افزودن به سبد'} onClick={() => variants.length ? go(`/products/${encodeURIComponent(product.slug)}`) : onAdd(product)}><Plus size={17} /></button>}
      </div>
    </div>
  </article>;
}

function ProductPlaceholder() {
  return <div className="product-placeholder"><Gem size={67} strokeWidth={1} /><span>MALEKI</span></div>;
}

function JewelryArtwork({ className = '' }) {
  return <div className={className} aria-hidden="true"><span className="auth-jewel-orbit auth-jewel-orbit-one" /><span className="auth-jewel-orbit auth-jewel-orbit-two" /><Gem size={214} strokeWidth={.65} /></div>;
}

function ProductSkeletons() { return <div className="product-grid">{Array.from({ length: 8 }, (_, index) => <div className="product-skeleton" key={index}><div /><span /><span /></div>)}</div>; }

function ProductPage({ slug, session, onAdd, toast, onNavigate }) {
  const [product, setProduct] = useState(null);
  const [comments, setComments] = useState([]);
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selected, setSelected] = useState(null);
  const [activeImage, setActiveImage] = useState('');
  const [comment, setComment] = useState('');

  useEffect(() => {
    let active = true;
    setLoading(true); setSelected(null); setError('');
    const previewProduct = DEMO_PREVIEW_MODE ? DEMO_PRODUCTS.find((item) => item.slug === slug) : null;
    if (previewProduct) {
      setProduct(previewProduct); setActiveImage(productImage(previewProduct)); setComments([]);
      setRecommendations(DEMO_PRODUCTS.filter((item) => item.slug !== previewProduct.slug));
      document.title = `${previewProduct.name} | MALEKI`;
      setLoading(false);
      return () => { active = false; };
    }
    api(`/products/slug/${encodeURIComponent(slug)}`, { auth: false }).then(async (item) => {
      const [commentResult, categoryResult, brandResult] = await Promise.allSettled([
        api(`/products/${item.id}/comments`, { auth: false }),
        item.category_id ? api(`/products?page=1&limit=8&category_id=${item.category_id}`, { auth: false }) : Promise.resolve({ data: [] }),
        item.brand_id ? api(`/products?page=1&limit=8&brand_id=${item.brand_id}`, { auth: false }) : Promise.resolve({ data: [] }),
      ]);
      if (!active) return;
      setProduct(item); setActiveImage(productImage(item));
      setComments(commentResult.status === 'fulfilled' ? (commentResult.value.data || commentResult.value || []) : []);
      const pool = [...(categoryResult.status === 'fulfilled' ? categoryResult.value.data || [] : []), ...(brandResult.status === 'fulfilled' ? brandResult.value.data || [] : [])];
      setRecommendations([...new Map(pool.filter((candidate) => String(candidate.id) !== String(item.id)).map((candidate) => [candidate.id, candidate])).values()].slice(0, 6));
      document.title = `${item.name} | MALEKI`;
    }).catch((failure) => { if (active) setError(failure.message); }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [slug]);

  if (loading) return <div className="page-width loading-view"><span className="spinner" />در حال آماده‌کردن انتخابت…</div>;
  if (error || !product) return <div className="page-width section-space"><ErrorPanel message={error || 'محصول پیدا نشد.'} /></div>;
  const variants = product.variants || [];
  const images = [...new Set([...(product.image_urls || []), product.image_url].filter(Boolean))];
  const source = selected || product;
  const price = activePrice(product, selected);
  const percent = discountPercent(source);
  const quickAttributes = Object.entries(product.attributes || {}).filter(([, value]) => String(value || '').trim()).slice(0, 4);
  const chooseVariant = (variant) => { setSelected(variant); setActiveImage(variant.image_url || productImage(product)); };
  const submitComment = async (event) => {
    event.preventDefault();
    if (product?.demo_only) return;
    if (!session) { onNavigate('/login'); return; }
    try { await api(`/products/${product.id}/comments`, { method: 'POST', body: { body: comment } }); setComment(''); toast('دیدگاهت پس از بررسی نمایش داده می‌شود.'); }
    catch (failure) { toast(failure.message, 'error'); }
  };

  return <div className="page-width product-page">
    {product.demo_only && <DemoPreviewNotice />}
    <div className="breadcrumbs"><a href={routeHref('/shop')}>فروشگاه</a><ArrowLeft size={13} /><span>{product.category_name || 'زیورآلات'}</span><ArrowLeft size={13} /><b>{product.name}</b></div>
    <section className="product-detail-grid">
      <div className="product-gallery"><div className="product-gallery-main">{activeImage ? <img className={product.demo_only ? 'demo-product-art' : ''} key={activeImage} src={activeImage} alt={product.name} onError={() => setActiveImage('')} /> : <ProductPlaceholder />}
        {percent > 0 && <span className="gallery-discount">{fa(percent)}٪ تخفیف</span>}</div>
        {quickAttributes.length > 0 && <ul className="product-quick-specs product-quick-specs-mobile" aria-label="ویژگی‌های کوتاه محصول">{quickAttributes.map(([key, value]) => <li key={key}><b>{key}</b><span>{value}</span></li>)}</ul>}
        {images.length > 1 && <div className="product-thumbnails">{images.map((image, index) => <button key={image} className={activeImage === image ? 'thumbnail thumbnail-active' : 'thumbnail'} onClick={() => setActiveImage(image)} aria-label={`تصویر ${fa(index + 1)}`}><img src={image} alt="" loading="lazy" /></button>)}</div>}
        {product.description && <section className="product-description-under"><span className="eyebrow">جزئیات محصول</span><RichTextBlock value={product.description} /></section>}
      </div>
      <div className="product-info"><div className="product-brand-line"><span>{product.brand_name || 'برند منتخب'}</span><span className="rating">{fa(comments.length)} نظر ثبت‌شده</span></div>
        <h1>{product.name}</h1>
        <ul className="product-quick-specs product-quick-specs-desktop" aria-label="ویژگی‌های کوتاه محصول">{quickAttributes.map(([key, value]) => <li key={key}><b>{key}</b><span>{value}</span></li>)}</ul>
        <div className="detail-price">{percent > 0 && <del>{money(source.price)}</del>}<strong>{money(price)}</strong>{percent > 0 && <span className="discount-badge">{fa(percent)}٪ تخفیف</span>}</div>
        {variants.length > 0 && <div className="size-picker"><div className="size-picker-heading"><strong>انتخاب {product.variant_label || 'گزینه'}</strong></div>
          <div className="size-grid">{variants.map((variant) => <button key={variant.id} type="button" className={`size-option ${selected?.id === variant.id ? 'size-selected' : ''}`} disabled={variant.stock <= 0} onClick={() => chooseVariant(variant)}>
            <span>{variant.name}</span>{variant.stock <= 0 && <small>تمام شد</small>}</button>)}</div>
          {selected && <div className="selected-size-price">{selected.discount_price > 0 && <del>{money(selected.price)}</del>}<b>{money(activePrice(product, selected))}</b>{selected.stock <= 5 && selected.stock > 0 && <span className="stock-warning">تنها {fa(selected.stock)} عدد باقی‌مانده</span>}</div>}
        </div>}
        <div className="detail-actions">{product.demo_only ? <div className="demo-purchase-note">این محصول فقط برای بررسی ظاهر فروشگاه اضافه شده و قابل خرید نیست.</div> : <button className="button button-red button-wide" disabled={!product.stock || (variants.length > 0 && !selected) || (selected && selected.stock <= 0)} onClick={() => onAdd(product, selected)}><ShoppingBag size={18} />افزودن به سبد خرید</button>}</div>
        <div className="product-guarantees"><span><ShieldCheck size={17} /> مشخصات محصول</span><span><Truck size={17} /> روش‌های ارسال در پرداخت</span><span><CircleHelp size={17} /> پشتیبانی از حساب کاربری</span></div>
      </div>
    </section>
    <section className="detail-reviews"><div className="review-panel panel-light"><div className="review-heading"><div><span className="eyebrow">نظر خریداران</span><h2>تجربه‌ها</h2></div><span className="rating rating-large">{fa(comments.length)} نظر</span></div>
        {comments.length ? comments.map((item) => <article className="review-item" key={item.id}><div><b>{item.author_name || 'خریدار فروشگاه'}</b><time>{shortDate(item.created_at)}</time></div><p>{item.body}</p></article>) : <p className="muted">هنوز نظری ثبت نشده؛ اولین نفر باش.</p>}
        {product.demo_only ? <div className="demo-purchase-note">ثبت دیدگاه برای محصول آزمایشی فعال نیست.</div> : <form className="review-form" onSubmit={submitComment}><textarea value={comment} onChange={(event) => setComment(event.target.value)} maxLength={2000} placeholder="تجربه‌ات از این مدل را بنویس…" required /><button className="button button-dark" type="submit"><Send size={16} />ثبت دیدگاه</button></form>}
      </div></section>
    {recommendations.length > 0 && <section className="related-section"><SectionTitle eyebrow="شاید این‌ها هم به دلت بنشینند" title="مدل‌های پیشنهادی" /><div className="product-grid">{recommendations.map((item, index) => <ProductCard key={item.id} product={item} index={index} onAdd={onAdd} />)}</div></section>}
  </div>;
}

function SectionTitle({ eyebrow, title, action }) { return <div className="section-title"><div><span className="eyebrow">{eyebrow}</span><h2>{title}</h2></div>{action}</div>; }
function ErrorPanel({ message }) { return <div className="error-panel"><CircleHelp size={20} /><span>{message}</span></div>; }
function EmptyState({ icon: Icon = Package, title, body, action }) { return <div className="empty-state"><span className="empty-icon"><Icon size={26} /></span><h2>{title}</h2><p>{body}</p>{action}</div>; }

function safeRichTextLink(rawValue) {
  let value = String(rawValue || "").trim();
  if (!value || /\s|[<>"]/u.test(value) || value.startsWith("//")) return "";
    if (/^(https?:\/\/|mailto:|tel:)/i.test(value)) return value;
  if (value.startsWith("/") || value.startsWith("#")) return value;
  if (/^[\w.-]+\.[a-z]{2,}(?::\d+)?(?:[/?#][^\s]*)?$/i.test(value)) return "https://" + value;
  return "";
}

function sanitizePastedRichText(rawHTML) {
  const parsed = new DOMParser().parseFromString(String(rawHTML || ""), "text/html");
  const allowed = new Set(["a", "b", "blockquote", "br", "div", "em", "h2", "h3", "h4", "hr", "i", "li", "ol", "p", "s", "strike", "strong", "sub", "sup", "u", "ul"]);
  const discarded = new Set(["iframe", "object", "script", "style", "svg", "math", "template"]);
  const clean = (node) => {
    if (node.nodeType === Node.TEXT_NODE) return document.createTextNode(node.nodeValue || "");
    if (node.nodeType !== Node.ELEMENT_NODE) return document.createDocumentFragment();
    const tag = node.tagName.toLowerCase();
    if (discarded.has(tag)) return document.createDocumentFragment();
    const children = document.createDocumentFragment();
    for (const child of node.childNodes) {
      const safe = clean(child);
      if (safe) children.append(safe);
    }
    if (!allowed.has(tag)) return children;
    const element = document.createElement(tag);
    if (tag === "a") {
      const href = safeRichTextLink(node.getAttribute("href"));
      if (href) element.setAttribute("href", href);
      const title = node.getAttribute("title");
      if (title && title.length <= 200) element.setAttribute("title", title);
      if (node.getAttribute("target") === "_blank") {
        element.setAttribute("target", "_blank");
        element.setAttribute("rel", "noopener noreferrer");
      }
    }
    element.append(children);
    return element;
  };
  const result = document.createElement("div");
  for (const child of parsed.body.childNodes) {
    const safe = clean(child);
    if (safe) result.append(safe);
  }
  return result.innerHTML;
}

function RichTextEditor({ label, value, onChange, placeholder = "متن را اینجا بنویس…", minHeight = 210 }) {
  const editor = useRef(null);
  const linkInput = useRef(null);
  const savedRange = useRef(null);
  const syncedValue = useRef(value || "");
  const mounted = useRef(false);
  const [blockFormat, setBlockFormat] = useState("p");
  const [linkOpen, setLinkOpen] = useState(false);
  const [linkURL, setLinkURL] = useState("");
  const [characterCount, setCharacterCount] = useState(() => richTextPlainText(value || "").length);

  useEffect(() => {
    if (!editor.current) return;
    const next = value || "";
    if (!mounted.current || next !== syncedValue.current) {
      if (editor.current.innerHTML !== next) editor.current.innerHTML = next;
      syncedValue.current = next;
      mounted.current = true;
      setCharacterCount(richTextPlainText(next).length);
    }
  }, [value]);

  const saveSelection = () => {
    const selection = window.getSelection();
    if (!editor.current || !selection?.rangeCount) return;
    const range = selection.getRangeAt(0);
    if (!editor.current.contains(range.commonAncestorContainer)) return;
    savedRange.current = range.cloneRange();
    let node = selection.anchorNode;
    if (node?.nodeType === Node.TEXT_NODE) node = node.parentElement;
    let format = "p";
    while (node && node !== editor.current) {
      const tag = node.tagName?.toLowerCase();
      if (["p", "h2", "h3", "h4", "blockquote"].includes(tag)) { format = tag; break; }
      node = node.parentNode;
    }
    setBlockFormat(format);
  };

  const restoreSelection = () => {
    if (!editor.current) return false;
    editor.current.focus();
    if (!editor.current.childNodes.length) editor.current.innerHTML = "<p><br></p>";
    const range = savedRange.current;
    if (range && editor.current.contains(range.commonAncestorContainer)) {
      const selection = window.getSelection();
      selection.removeAllRanges();
      // A caret saved while the editor was empty points at the editor root.
      // Once its first paragraph exists, move that caret inside the paragraph.
      if (range.startContainer === editor.current) {
        const firstBlock = editor.current.firstElementChild;
        const nextRange = document.createRange();
        nextRange.selectNodeContents(firstBlock || editor.current);
        nextRange.collapse(true);
        selection.addRange(nextRange);
      } else {
        selection.addRange(range);
      }
    } else {
      const selection = window.getSelection();
      const nextRange = document.createRange();
      const last = editor.current.lastChild;
      nextRange.selectNodeContents(last || editor.current);
      nextRange.collapse(false);
      selection.removeAllRanges();
      selection.addRange(nextRange);
    }
    return true;
  };

  const syncContent = () => {
    if (!editor.current) return;
    const html = editor.current.innerHTML;
    syncedValue.current = html;
    setCharacterCount(richTextPlainText(html).length);
    onChange(html);
  };

  const execute = (command, argument) => {
    restoreSelection();
    const valueArgument = command === "formatBlock" ? "<" + argument + ">" : argument;
    document.execCommand(command, false, valueArgument);
    syncContent();
    saveSelection();
  };

  const openLinkEditor = () => {
    saveSelection();
    const selection = window.getSelection();
    let node = selection?.anchorNode;
    if (node?.nodeType === Node.TEXT_NODE) node = node.parentElement;
    const anchor = node?.closest?.("a");
    setLinkURL(anchor?.getAttribute("href") || "");
    setLinkOpen(true);
    window.requestAnimationFrame(() => linkInput.current?.focus());
  };

  const applyLink = (event) => {
    event.preventDefault();
    const href = safeRichTextLink(linkURL);
    if (!href) return;
    restoreSelection();
    const selection = window.getSelection();
    if (selection?.rangeCount) {
      const range = selection.getRangeAt(0);
      let node = range.startContainer;
      if (node.nodeType === Node.TEXT_NODE) node = node.parentElement;
      const existingLink = node?.closest?.("a");
      let anchor;
      if (existingLink && editor.current.contains(existingLink)) {
        anchor = existingLink;
        anchor.setAttribute("href", href);
      } else {
        anchor = document.createElement("a");
        anchor.setAttribute("href", href);
        if (range.collapsed) {
          anchor.textContent = href;
          range.insertNode(anchor);
        } else {
          anchor.appendChild(range.extractContents());
          range.insertNode(anchor);
        }
      }
      range.setStartAfter(anchor);
      range.collapse(true);
      selection.removeAllRanges();
      selection.addRange(range);
    } else {
    }
    setLinkOpen(false);
    syncContent();
    saveSelection();
  };

  const handlePaste = (event) => {
    event.preventDefault();
    const html = event.clipboardData?.getData("text/html");
    const plain = event.clipboardData?.getData("text/plain") || "";
    const safeHTML = html
      ? sanitizePastedRichText(html)
      : plain.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/\r\n?|\n/g, "<br>");
    restoreSelection();
    document.execCommand("insertHTML", false, safeHTML);
    syncContent();
    saveSelection();
  };

  const tool = (title, face, command, argument) => <button type="button" title={title} aria-label={title} onMouseDown={(event) => { saveSelection(); event.preventDefault(); }} onClick={() => execute(command, argument)}>{face}</button>;
  return <div className="rich-editor field">
    <span className="field-label">{label}</span>
    <div className="rich-toolbar" role="toolbar" aria-label={"ابزارهای " + label}>
      <div className="rich-tool-group">
        <select aria-label="نوع پاراگراف یا تیتر" value={blockFormat} onMouseDown={saveSelection} onChange={(event) => execute("formatBlock", event.target.value)}>
          <option value="p">متن عادی</option><option value="h2">تیتر بزرگ</option><option value="h3">تیتر میانی</option><option value="h4">تیتر کوچک</option><option value="blockquote">نقل‌قول</option>
        </select>
      </div>
      <div className="rich-tool-group">
        {tool("پررنگ", "B", "bold")}{tool("ایتالیک", "I", "italic")}{tool("زیرخط", "U", "underline")}{tool("خط‌خورده", "S̶", "strikeThrough")}
        {tool("زیرنویس", "x₂", "subscript")}{tool("بالانویس", "x²", "superscript")}
      </div>
      <div className="rich-tool-group">
        {tool("فهرست نشانه‌دار", "• فهرست", "insertUnorderedList")}{tool("فهرست شماره‌دار", "۱. فهرست", "insertOrderedList")}
        {tool("خط جداکننده", "― خط", "insertHorizontalRule")}
      </div>
      <div className="rich-tool-group">
        <button type="button" title="افزودن یا ویرایش لینک" onMouseDown={(event) => { saveSelection(); event.preventDefault(); }} onClick={openLinkEditor}>↗ لینک</button>
        {tool("حذف لینک", "حذف لینک", "unlink")}
        {tool("واگرد", "↶", "undo")}{tool("ازنو", "↷", "redo")}{tool("پاک‌کردن قالب متن", "Tx", "removeFormat")}
      </div>
    </div>
    {linkOpen && <form className="rich-link-panel" onSubmit={applyLink}>
      <label htmlFor={"rich-link-" + label}>نشانی لینک</label><input ref={linkInput} id={"rich-link-" + label} dir="ltr" type="text" autoComplete="url" value={linkURL} onChange={(event) => setLinkURL(event.target.value)} placeholder="https://example.com" />
      <button className="button button-red button-small" type="submit">ثبت لینک</button><button className="button button-outline button-small" type="button" onMouseDown={saveSelection} onClick={() => setLinkOpen(false)}>انصراف</button>
      <small>می‌توانی ابتدا بخشی از متن را انتخاب کنی؛ در غیر این صورت نشانی به متن لینک تبدیل می‌شود.</small>
    </form>}
    <div ref={editor} className="rich-editor-surface" contentEditable suppressContentEditableWarning role="textbox" aria-label={label} aria-multiline="true" data-placeholder={placeholder} style={{ minHeight }}
      onInput={() => { syncContent(); saveSelection(); }} onKeyUp={saveSelection} onMouseUp={saveSelection} onFocus={saveSelection} onPaste={handlePaste} />
    <div className="rich-editor-footer"><small>تیتر، قالب‌بندی، فهرست و لینک برای متن بلند</small><small>{fa(characterCount)} نویسه</small></div>
  </div>;
}

function RichTextBlock({ value, className = '' }) {
  if (!value) return null;
  return <div className={`rich-text-content ${className}`} dangerouslySetInnerHTML={{ __html: value }} />;
}

function richTextPlainText(value) {
  const documentValue = new DOMParser().parseFromString(value || '', 'text/html');
  return (documentValue.body.textContent || '').replace(/\s+/g, ' ').trim();
}

function SiteContentPage({ page, session, toast, onNavigate }) {
  const [data, setData] = useState(null);
  const [error, setError] = useState('');
  useEffect(() => {
    let active = true;
    api('/site-content', { auth: false }).then((value) => { if (active) setData(value); }).catch((failure) => { if (active) setError(failure.message); });
    return () => { active = false; };
  }, []);
  useEffect(() => {
    const titles = { contact: 'تماس با ما', about: 'درباره MALEKI', terms: 'قوانین و مقررات' };
    const title = titles[page] || 'MALEKI';
    document.title = `${title} | MALEKI`;
    let meta = document.querySelector('meta[name="description"]');
    if (!meta) { meta = document.createElement('meta'); meta.name = 'description'; document.head.append(meta); }
    const copy = page === 'about' ? data?.about_content : page === 'terms' ? data?.terms_content : `${data?.contact_address || ''} ${data?.contact_hours || ''}`;
    meta.content = `${richTextPlainText(copy)} ${title}`.trim().slice(0, 320);
  }, [page, data]);
  if (error) return <div className="page-width section-space"><ErrorPanel message={error} /></div>;
  if (!data) return <div className="page-width loading-view"><span className="spinner" />در حال بارگذاری…</div>;
  if (page === 'contact') return <main className="page-width info-page contact-page">
    <header className="info-page-heading"><span className="eyebrow">MALEKI · ارتباط</span><h1>تماس با ما</h1><p>برای پرسش، پیگیری یا همراهی، از راهی که برایت راحت‌تر است پیام بده.</p></header>
    <div className="contact-cards">
      <article className="contact-card"><span className="contact-card-mark">☎</span><small>شمارهٔ تماس</small>{data.contact_phone ? <a dir="ltr" href={`tel:${data.contact_phone.replace(/[^+\d]/g, '')}`}>{data.contact_phone}</a> : <b>هنوز ثبت نشده</b>}</article>
      <article className="contact-card"><span className="contact-card-mark">@</span><small>ایمیل</small>{data.contact_email ? <a dir="ltr" href={`mailto:${data.contact_email}`}>{data.contact_email}</a> : <b>هنوز ثبت نشده</b>}</article>
      <article className="contact-card"><span className="contact-card-mark">⌖</span><small>نشانی</small><b>{data.contact_address || 'هنوز ثبت نشده'}</b></article>
      <article className="contact-card"><span className="contact-card-mark">◷</span><small>ساعت پاسخ‌گویی</small><b>{data.contact_hours || 'از طریق تیکت پیام بگذارید'}</b>{data.instagram_url && <a dir="ltr" href={data.instagram_url} target="_blank" rel="noreferrer">Instagram ↗</a>}</article>
    </div>
    <section className="contact-ticket-section"><div><span className="eyebrow">پشتیبانی MALEKI</span><h2>نیاز به راهنمایی داری؟</h2><p>پیامت را به شکل تیکت بفرست تا در حساب کاربری‌ات پیگیری شود.</p></div>{session ? <TicketsPanel toast={toast} /> : <div className="contact-login-card"><p>برای ثبت و پیگیری تیکت، ابتدا وارد حساب کاربری شو.</p><button className="button button-dark" onClick={() => onNavigate('/login')}>ورود به حساب <ArrowLeft size={15} /></button></div>}</section>
  </main>;
  const isAbout = page === 'about';
  return <main className="page-width info-page long-form-page"><header className="info-page-heading"><span className="eyebrow">MALEKI · {isAbout ? 'داستان ما' : 'اطلاعات حقوقی'}</span><h1>{isAbout ? 'درباره ما' : 'قوانین و مقررات'}</h1></header><section className="long-form-panel"><RichTextBlock value={isAbout ? data.about_content : data.terms_content} />{!(isAbout ? data.about_content : data.terms_content) && <p className="muted">این بخش هنوز توسط مدیر فروشگاه تکمیل نشده است.</p>}</section></main>;
}

function LoginPage({ onLogin, admin = false }) {
  const [step, setStep] = useState('phone');
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [showRegistration, setShowRegistration] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const requestCode = async (event) => {
    event.preventDefault(); setBusy(true); setError('');
    try { await api('/auth/otp/request', { method: 'POST', auth: false, body: { phone: phone.trim() } }); setStep('code'); }
    catch (failure) { setError(failure.message); }
    finally { setBusy(false); }
  };
  const verifyCode = async (event) => {
    event.preventDefault(); setBusy(true); setError('');
    const body = { phone: phone.trim(), code: code.trim() };
    if (showRegistration) {
      body.first_name = firstName.trim(); body.last_name = lastName.trim(); body.username = username.trim();
      if (email.trim()) body.email = email.trim();
    }
    try {
      const login = await api('/auth/otp/verify', { method: 'POST', auth: false, body });
      onLogin({ token: login.access_token, expiresAt: login.expires_at, user: login.user });
    } catch (failure) {
      if (failure.status === 422 && (failure.details?.first_name || failure.details?.last_name || failure.details?.username)) {
        setShowRegistration(true);
        setError('برای ساخت حساب، نام، نام خانوادگی و نام کاربری را تکمیل کن.');
      } else setError(failure.message);
    }
    finally { setBusy(false); }
  };

  return <section className="auth-page page-width"><div className="auth-art"><JewelryArtwork className="auth-jewelry-art" /><div className="auth-art-copy"><span>MALEKI · JEWELRY GALLERY</span><h2>زیبایی،<br />ماندگار.</h2></div></div>
    <div className="auth-panel"><a className="back-link" href={routeHref('/')}><ArrowRight size={16} /> بازگشت به فروشگاه</a><div className="auth-title"><span className="brand-mark"><Gem size={21} /></span><span className="eyebrow">{admin ? 'ورود مدیر' : 'به MALEKI خوش آمدی'}</span><h1>{step === 'phone' ? 'ورود یا ثبت‌نام' : 'کد تایید را وارد کن'}</h1><p>{step === 'phone' ? 'با شماره موبایل، سریع و امن وارد شو.' : `کد ارسال‌شده به ${phone} را وارد کن.`}</p></div>
      <form className="form-stack" onSubmit={step === 'phone' ? requestCode : verifyCode}>
        {step === 'phone' ? <label className="field"><span>شماره موبایل</span><input dir="ltr" inputMode="tel" autoComplete="tel" placeholder="0912 345 6789" value={phone} onChange={(event) => setPhone(event.target.value)} required /></label> : <>
          <label className="field"><span>کد یک‌بارمصرف</span><input dir="ltr" inputMode="numeric" autoComplete="one-time-code" placeholder="کد ۶ رقمی" maxLength={6} value={code} onChange={(event) => setCode(event.target.value)} required /></label>
          {showRegistration && <><div className="signup-fields"><label className="field"><span>نام</span><input autoComplete="given-name" value={firstName} onChange={(event) => setFirstName(event.target.value)} required /></label><label className="field"><span>نام خانوادگی</span><input autoComplete="family-name" value={lastName} onChange={(event) => setLastName(event.target.value)} required /></label></div>
            <label className="field"><span>نام کاربری</span><input dir="ltr" autoComplete="username" minLength={3} maxLength={30} pattern="[A-Za-z0-9_]{3,30}" placeholder="مثلاً kicks_user" value={username} onChange={(event) => setUsername(event.target.value)} required /><small className="field-hint">۳ تا ۳۰ حرف انگلیسی، عدد یا زیرخط</small></label>
            <label className="field"><span>ایمیل <small>(اختیاری)</small></span><input dir="ltr" type="email" autoComplete="email" placeholder="name@example.com" value={email} onChange={(event) => setEmail(event.target.value)} /></label></>}
        </>}
        {error && <p className="form-error">{error}</p>}
        <button className="button button-red button-wide" disabled={busy}>{busy ? <span className="spinner spinner-small" /> : step === 'phone' ? 'دریافت کد تایید' : showRegistration ? 'ساخت حساب و ورود' : 'ادامه'}<ArrowLeft size={17} /></button>
      </form>
      {step === 'code' && <button className="text-link auth-resend" onClick={() => { setStep('phone'); setCode(''); setError(''); }}>شماره را اشتباه وارد کردی؟ <span>ویرایش شماره</span></button>}
      <p className="auth-legal">با ادامه، <a href="/blog">قوانین و حریم خصوصی MALEKI</a> را می‌پذیری.</p>
    </div></section>;
}

function AccountPage({ route, session, toast, onLogout }) {
  const [profile, setProfile] = useState(session.user);
  const [loading, setLoading] = useState(false);
  useEffect(() => { api('/me').then(setProfile).catch(() => {}); }, []);
  const path = route.path;
  const ticketMatch = path.match(/^\/account\/tickets\/(\d+)$/);
  const section = ticketMatch ? 'tickets' : path.split('/')[2] || 'overview';
  const tabs = [
    ['/account', 'اطلاعات حساب', UserRound], ['/account/orders', 'سفارش‌های من', Package],
    ['/account/addresses', 'آدرس‌های من', MapPin], ['/account/tickets', 'پشتیبانی', CircleHelp],
  ];
  return <div className="page-width account-page-layout"><aside className="account-sidebar"><div className="account-user"><span className="account-avatar">{(profile.first_name || 'ک').slice(0, 1)}</span><span><b>{profile.first_name} {profile.last_name}</b><small dir="ltr">{profile.phone}</small></span></div>
    <nav>{tabs.map(([href, label, Icon]) => <a key={href} className={(href === '/account' ? section === 'overview' : path.startsWith(href)) ? 'account-tab active' : 'account-tab'} href={routeHref(href)}><Icon size={17} />{label}<ArrowLeft size={14} className="tab-arrow" /></a>)}</nav>
    <button className="account-logout" onClick={onLogout}><LogOut size={16} />خروج از حساب</button></aside>
    <section className="account-workspace"><div className="account-page-heading"><span className="eyebrow">حساب MALEKI</span><h1>{section === 'overview' ? 'سلام، ' + (profile.first_name || 'دوست عزیز') : section === 'orders' ? 'سفارش‌های من' : section === 'addresses' ? 'نشانی‌های من' : ticketMatch ? 'گفت‌وگوی پشتیبانی' : 'تیکت‌های پشتیبانی'}</h1><p>مدیریت خریدها و اطلاعات حساب از یک جا.</p></div>
      {section === 'overview' && <ProfileOverview profile={profile} />}
      {section === 'orders' && <OrdersPanel toast={toast} />}
      {section === 'addresses' && <AddressesPanel toast={toast} />}
      {section === 'tickets' && (ticketMatch ? <TicketConversation id={ticketMatch[1]} toast={toast} /> : <TicketsPanel toast={toast} />)}
    </section>
  </div>;
}

function ProfileOverview({ profile }) {
  return <div className="profile-overview"><div className="profile-card"><div className="profile-card-top"><span className="account-avatar profile-avatar">{(profile.first_name || 'ک').slice(0, 1)}</span><div><span className="eyebrow">حساب کاربری</span><h2>{profile.first_name} {profile.last_name}</h2></div><span className="verified"><Check size={14} /> فعال</span></div>
    <div className="profile-data"><div><small>شماره موبایل</small><b dir="ltr">{profile.phone}</b></div><div><small>ایمیل</small><b dir="ltr">{profile.email || 'ثبت نشده'}</b></div><div><small>تاریخ عضویت</small><b>{shortDate(profile.created_at)}</b></div><div><small>شناسه کاربری</small><b>#{fa(profile.id)}</b></div></div></div>
    <div className="account-quick-grid"><a href={routeHref('/account/orders')}><Package size={19} /><b>سفارش‌ها</b><span>پیگیری خرید و وضعیت ارسال</span><ArrowLeft size={15} /></a><a href={routeHref('/account/addresses')}><MapPin size={19} /><b>نشانی‌ها</b><span>مدیریت نشانی‌های تحویل</span><ArrowLeft size={15} /></a><a href={routeHref('/account/tickets')}><CircleHelp size={19} /><b>پشتیبانی</b><span>گفت‌وگو با تیم فروشگاه</span><ArrowLeft size={15} /></a></div></div>;
}

function OrdersPanel({ toast }) {
  const [orders, setOrders] = useState([]); const [loading, setLoading] = useState(true);
  useEffect(() => { api('/orders?page=1&limit=30').then((data) => setOrders(data.data || [])).catch((failure) => toast(failure.message, 'error')).finally(() => setLoading(false)); }, []);
  const pay = async (order) => { try { const payment = await api(`/orders/${order.id}/pay`, { method: 'POST' }); go(`/pay/${payment.authority}`); } catch (failure) { toast(failure.message, 'error'); } };
  if (loading) return <LoadingState />;
  return orders.length ? <div className="orders-list">{orders.map((order) => <article className="order-card" key={order.id}><div className="order-card-head"><div><small>سفارش #{fa(order.id)}</small><time>{dateTime(order.created_at)}</time></div><span className={`status-pill status-${order.status}`}>{ORDER_LABELS[order.status] || order.status}</span></div>
    <div className="order-items-preview">{order.items.map((item, index) => <div key={`${item.product_id}-${index}`}><span>{item.product_name}{item.variant_name ? ` · ${item.variant_name}` : ''}</span><small>{fa(item.quantity)} عدد × {money(item.unit_price)}</small></div>)}</div>
    <div className="order-card-bottom"><span>مبلغ سفارش <b>{money(order.total_amount)}</b></span><a className="text-link tracking-link" href={routeHref(`/track/${encodeURIComponent(order.tracking_code)}`)}>پیگیری {order.tracking_code} <ArrowUpLeft size={14} /></a>{order.status === 'awaiting_payment' && <button className="button button-red button-small" onClick={() => pay(order)}>پرداخت سفارش <ArrowLeft size={15} /></button>}</div></article>)}</div>
    : <EmptyState icon={Package} title="هنوز سفارشی نداری" body="مدل بعدی را پیدا کن و از همین‌جا وضعیت خریدت را پیگیری کن." action={<a className="button button-red" href={routeHref('/shop')}>دیدن زیورآلات‌ها <ArrowLeft size={16} /></a>} />;
}

function AddressesPanel({ toast }) {
  const [addresses, setAddresses] = useState([]); const [form, setForm] = useState(emptyAddress); const [editing, setEditing] = useState(null); const [busy, setBusy] = useState(false);
  const [manualCity, setManualCity] = useState(false);
  const provincePlaces = IRAN_LOCATIONS[form.province] || { cities: [], counties: [] };
  const load = () => api('/addresses').then(setAddresses).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, []);
  const update = (event) => {
    const { name, value } = event.target;
    if (name === 'province') { setManualCity(false); setForm((current) => ({ ...current, province: value, city: '' })); return; }
    if (name === 'city-choice') { setManualCity(value === '__manual__'); if (value !== '__manual__') setForm((current) => ({ ...current, city: value })); else setForm((current) => ({ ...current, city: '' })); return; }
    setForm((current) => ({ ...current, [name]: value }));
  };
  const submit = async (event) => {
    event.preventDefault(); setBusy(true);
    try { await api(editing ? `/addresses/${editing}` : '/addresses', { method: editing ? 'PUT' : 'POST', body: form }); setForm(emptyAddress); setEditing(null); setManualCity(false); await load(); toast('آدرس ذخیره شد.'); }
    catch (failure) { toast(failure.message, 'error'); } finally { setBusy(false); }
  };
  const edit = (address) => { const places = IRAN_LOCATIONS[address.province] || { cities: [], counties: [] }; setEditing(address.id); setForm(Object.fromEntries(Object.keys(emptyAddress).map((key) => [key, address[key] || '']))); setManualCity(![...places.cities, ...places.counties].includes(address.city)); window.scrollTo({ top: 0, behavior: 'smooth' }); };
  const remove = async (address) => { if (!confirm('این آدرس حذف شود؟')) return; try { await api(`/addresses/${address.id}`, { method: 'DELETE' }); await load(); toast('آدرس حذف شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  const makeDefault = async (address) => { try { await api(`/addresses/${address.id}/default`, { method: 'PATCH' }); await load(); toast('آدرس پیش‌فرض تغییر کرد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className="address-manager"><form className="form-card" onSubmit={submit}><div className="form-card-heading"><div><span className="eyebrow">نشانی تحویل</span><h2>{editing ? 'ویرایش آدرس' : 'افزودن آدرس تازه'}</h2></div><MapPin size={19} /></div>
    <div className="form-grid">{[['receiver_name', 'نام گیرنده'], ['receiver_phone', 'شماره تماس']].map(([name, label]) => <label className="field" key={name}><span>{label}</span><input name={name} value={form[name]} onChange={update} required dir={name.includes('phone') ? 'ltr' : undefined} /></label>)}
      <label className="field"><span>استان</span><select name="province" value={form.province} onChange={update} required><option value="">انتخاب استان</option>{IRAN_PROVINCES.map((province) => <option key={province}>{province}</option>)}</select></label>
      <label className="field"><span>شهر / شهرستان</span><select name="city-choice" value={manualCity ? '__manual__' : form.city} onChange={update} required disabled={!form.province}><option value="">{form.province ? 'جستجو و انتخاب محل' : 'ابتدا استان را انتخاب کنید'}</option><optgroup label="شهرها">{provincePlaces.cities.map((city, index) => <option key={`${city}-${index}`}>{city}</option>)}</optgroup><optgroup label="شهرستان‌ها">{provincePlaces.counties.map((county, index) => <option key={`${county}-${index}`}>{county}</option>)}</optgroup><option value="__manual__">نام در فهرست نیست…</option></select></label>
      {manualCity && <label className="field"><span>نام شهر یا شهرستان</span><input name="city" value={form.city} onChange={update} required placeholder="نام محل سکونت" /></label>}
      <label className="field"><span>کد پستی</span><input name="postal_code" value={form.postal_code} onChange={update} required dir="ltr" /></label>
      <label className="field field-wide"><span>نشانی کامل</span><textarea name="address_line" rows={3} value={form.address_line} onChange={update} required /></label></div>
    <div className="form-actions"><button className="button button-dark" disabled={busy}>{editing ? 'ذخیره تغییرات' : 'ثبت آدرس'} <ArrowLeft size={15} /></button>{editing && <button type="button" className="button button-outline" onClick={() => { setEditing(null); setForm(emptyAddress); }}>انصراف</button>}</div></form>
    <div className="address-list">{addresses.map((address) => <article className="address-card" key={address.id}><div className="address-card-heading"><MapPin size={17} /><b>{address.receiver_name}</b>{address.is_default && <span className="verified">پیش‌فرض</span>}<div className="address-actions"><button aria-label="ویرایش" onClick={() => edit(address)}><PenIcon /></button><button aria-label="حذف" onClick={() => remove(address)}><Trash2 size={15} /></button></div></div><p>{address.province}، {address.city} · {address.address_line}</p><small>{address.receiver_phone} · کد پستی {address.postal_code}</small>{!address.is_default && <button className="text-link" onClick={() => makeDefault(address)}>انتخاب به‌عنوان پیش‌فرض</button>}</article>)}{!addresses.length && <p className="muted">هنوز نشانی‌ای ثبت نکردی.</p>}</div></div>;
}

const emptyAddress = { receiver_name: '', receiver_phone: '', province: '', city: '', address_line: '', postal_code: '' };
function PenIcon() { return <span className="pen-icon">✎</span>; }

function TicketsPanel({ toast }) {
  const [data, setData] = useState([]); const [subject, setSubject] = useState(''); const [body, setBody] = useState(''); const [busy, setBusy] = useState(false);
  const load = () => api('/tickets?page=1&limit=50').then((result) => setData(result.data || [])).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, []);
  const submit = async (event) => { event.preventDefault(); setBusy(true); try { const ticket = await api('/tickets', { method: 'POST', body: { subject, body } }); setSubject(''); setBody(''); await load(); go(`/account/tickets/${ticket.id}`); toast('درخواست پشتیبانی ثبت شد.'); } catch (failure) { toast(failure.message, 'error'); } finally { setBusy(false); } };
  return <div className="ticket-layout"><form className="form-card" onSubmit={submit}><div className="form-card-heading"><div><span className="eyebrow">ما اینجاییم</span><h2>چطور کمک کنیم؟</h2></div><CircleHelp size={20} /></div><label className="field"><span>موضوع</span><input value={subject} onChange={(event) => setSubject(event.target.value)} maxLength={200} required placeholder="مثلاً راهنمای انتخاب سایز" /></label><label className="field"><span>پیامت</span><textarea value={body} onChange={(event) => setBody(event.target.value)} maxLength={5000} rows={5} required placeholder="جزئیات را برایمان بنویس…" /></label><button className="button button-red" disabled={busy}>ارسال درخواست <Send size={16} /></button></form>
    <div className="ticket-list"><SectionTitle eyebrow="پیگیری پیام‌ها" title="درخواست‌های من" />{data.map((ticket) => <a className="ticket-row" href={routeHref(`/account/tickets/${ticket.id}`)} key={ticket.id}><span className="ticket-row-icon"><MessageIcon /></span><span className="ticket-row-copy"><b>{ticket.subject}</b><small>{ticket.last_message || 'بدون پیام'}</small></span><span className={`status-pill status-${ticket.status}`}>{ticket.status === 'answered' ? 'پاسخ داده شد' : ticket.status === 'closed' ? 'بسته' : 'در انتظار'}</span><ArrowLeft size={15} /></a>)}{!data.length && <p className="muted">درخواستی ثبت نشده است.</p>}</div></div>;
}

function TicketConversation({ id, toast }) {
  const [ticket, setTicket] = useState(null); const [body, setBody] = useState(''); const [busy, setBusy] = useState(false);
  const load = () => api(`/tickets/${id}`).then(setTicket).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, [id]);
  const submit = async (event) => { event.preventDefault(); setBusy(true); try { await api(`/tickets/${id}/messages`, { method: 'POST', body: { body } }); setBody(''); await load(); } catch (failure) { toast(failure.message, 'error'); } finally { setBusy(false); } };
  if (!ticket) return <LoadingState />;
  return <div className="conversation-card"><div className="conversation-heading"><a className="back-link" href={routeHref('/account/tickets')}><ArrowRight size={15} /> بازگشت به تیکت‌ها</a><span className={`status-pill status-${ticket.status}`}>{ticket.status === 'answered' ? 'پاسخ داده شد' : ticket.status === 'closed' ? 'بسته' : 'باز'}</span><h2>{ticket.subject}</h2></div>
    <div className="conversation-messages">{(ticket.messages || []).map((message) => <article key={message.id} className={`message-bubble ${message.sender_role === 'admin' ? 'message-support' : 'message-customer'}`}><div><b>{message.sender_role === 'admin' ? 'پشتیبانی MALEKI' : message.sender_name}</b><time>{dateTime(message.created_at)}</time></div><p>{message.body}</p></article>)}</div>
    {ticket.status !== 'closed' ? <form className="reply-form" onSubmit={submit}><textarea value={body} onChange={(event) => setBody(event.target.value)} maxLength={5000} required placeholder="پاسخت را بنویس…" /><button className="button button-red" disabled={busy}>ارسال پیام <Send size={16} /></button></form> : <p className="muted">این گفت‌وگو بسته شده است.</p>}</div>;
}

function MessageIcon() { return <Send size={15} />; }

function BlogPage({ path, onNavigate }) {
  const slug = path.startsWith('/blog/') ? decodeURIComponent(path.slice('/blog/'.length)) : '';
  const [data, setData] = useState(null); const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [listLoading, setListLoading] = useState(false);
  useEffect(() => {
    let active = true;
    if (slug) setData(null);
    setError('');
    if (!slug) setListLoading(true);
    const timer = window.setTimeout(() => {
      const params = new URLSearchParams({ page: '1', limit: '50' });
      if (search.trim() && !slug) params.set('q', search.trim());
      api(slug ? `/blog/posts/${encodeURIComponent(slug)}` : `/blog/posts?${params}`, { auth: false })
        .then((value) => { if (active) setData(value); })
        .catch((failure) => { if (active) setError(failure.message); })
        .finally(() => { if (active) setListLoading(false); });
    }, slug ? 0 : 250);
    return () => { active = false; window.clearTimeout(timer); };
  }, [slug, search]);
  useEffect(() => {
    if (!data) return;
    const post = slug ? data : null;
    document.title = post ? `${post.seo_title || post.title} | مجله MALEKI` : 'مجله MALEKI | راهنمای انتخاب زیورآلات';
    let description = document.querySelector('meta[name="description"]');
    if (!description) { description = document.createElement('meta'); description.name = 'description'; document.head.append(description); }
    description.content = post ? post.seo_description || post.summary : 'راهنمای انتخاب سایز، مراقبت از زیورآلات و تازه‌ترین مجموعه‌ها در مجله MALEKI.';
    let canonical = document.querySelector('link[rel="canonical"]');
    if (!canonical) { canonical = document.createElement('link'); canonical.rel = 'canonical'; document.head.append(canonical); }
    canonical.href = `${location.origin}${post ? `/blog/${post.slug}` : '/blog'}`;
  }, [data, slug]);
  if (error) return <div className="page-width section-space"><ErrorPanel message={error} /></div>;
  if (!data) return <div className="page-width loading-view"><span className="spinner" />در حال بارگذاری مجله…</div>;
  if (slug) {
    const wordCount = richTextPlainText(data.content).split(/\s+/).filter(Boolean).length;
    const readMinutes = Math.max(1, Math.ceil(wordCount / 180));
    return <article className="page-width blog-article-page">
      <a className="back-link" href="/blog"><ArrowRight size={15} /> بازگشت به مجله</a>
      {data.cover_image_url && <img className="blog-cover" src={data.cover_image_url} alt={data.title} />}
      <div className="blog-article-copy">
        <span className="eyebrow">مجلهٔ ملکی · راهنمای زیورآلات</span><h1>{data.title}</h1>
        <div className="blog-article-meta"><span>تیم تحریریهٔ ملکی</span><time>{shortDate(data.published_at)}</time><span><Clock3 size={14} /> {fa(readMinutes)} دقیقه مطالعه</span></div>
        {data.summary && <p className="blog-summary">{data.summary}</p>}
        <RichTextBlock value={data.content} className="blog-content" />
        <div className="article-cta"><span>مدل مناسب خودت را پیدا کردی؟</span><button className="button button-red" onClick={() => onNavigate('/shop')}>رفتن به فروشگاه <ArrowLeft size={16} /></button></div>
      </div>
    </article>;
  }
  const posts = data.data || [];
  return <div className="page-width magazine-page">
    <section className="magazine-hero"><span className="eyebrow"><Sparkles size={14} /> مجلهٔ ملکی</span><h1>زیورآلات خوب،<br /><em>انتخابی ماندگار.</em></h1><p>راهنمای انتخاب، نگهداری و شناخت زیورآلات؛ نوشته‌هایی برای انتخاب آگاهانه‌تر.</p><span className="magazine-mark">مجله</span></section>
    <div className="magazine-toolbar"><div><span className="eyebrow">دانستنی‌های گالری</span><h2>تازه از مجله</h2><p>{fa(data.total ?? posts.length)} نوشته برای الهام و راهنمایی</p></div>
      <label className="magazine-search"><Search size={17} /><input type="search" aria-label="جست‌وجو در مجله" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="عنوان یا موضوع مطلب…" /></label>
    </div>
    {error ? <ErrorPanel message={error} /> : posts.length ? <div className="blog-grid">{posts.map((post, index) => {
      const url = `/blog/${encodeURIComponent(post.slug)}`;
      const words = richTextPlainText(post.content || post.summary).split(/\s+/).filter(Boolean).length;
      const readMinutes = Math.max(1, Math.ceil(words / 180));
      return <article className={`blog-card ${index === 0 && !search.trim() ? 'blog-card-featured' : ''}`} key={post.id}>
        <a href={url} className="blog-card-image">{post.cover_image_url ? <img src={post.cover_image_url} alt={post.title} loading="lazy" /> : <ProductPlaceholder />}<span className="blog-card-arrow"><ArrowUpLeft size={17} /></span></a>
        <div className="blog-card-copy"><span className="eyebrow">راهنمای ملکی</span><h2><a href={url}>{post.title}</a></h2><p>{post.summary}</p>
          <div className="blog-card-meta"><time>{shortDate(post.published_at)}</time><span><Clock3 size={13} /> {fa(readMinutes)} دقیقه</span></div>
          <a className="text-link" href={url}>ادامهٔ مطلب <ArrowLeft size={14} /></a>
        </div>
      </article>;
    })}</div> : listLoading ? <div className="loading-view"><span className="spinner" />در حال جست‌وجو در مجله…</div>
      : <EmptyState icon={Search} title={search.trim() ? 'مطلبی پیدا نشد' : 'مجله در راه است'} body={search.trim() ? 'عبارت جست‌وجو را کوتاه‌تر یا متفاوت وارد کن.' : 'به‌زودی راهنماهای انتخاب و نگهداری از زیورآلات اینجا منتشر می‌شوند.'} action={search.trim() ? <button className="button button-outline" onClick={() => setSearch('')}>پاک‌کردن جست‌وجو</button> : <button className="button button-outline" onClick={() => onNavigate('/')}>دیدن فروشگاه</button>} />}
  </div>;
}

function PaymentPage({ authority, toast, onNavigate }) {
  const [payment, setPayment] = useState(null); const [busy, setBusy] = useState(false);
  useEffect(() => { api(`/payments/${encodeURIComponent(authority)}`, { auth: false }).then(setPayment).catch((failure) => toast(failure.message, 'error')); }, [authority]);
  const confirm = async (action) => { setBusy(true); try { setPayment(await api(`/payments/${encodeURIComponent(authority)}/confirm`, { method: 'POST', auth: false, body: { action } })); } catch (failure) { toast(failure.message, 'error'); } finally { setBusy(false); } };
  if (!payment) return <div className="page-width loading-view"><span className="spinner" />در حال اتصال به درگاه…</div>;
  const done = payment.status !== 'pending';
  return <div className="payment-page"><div className="payment-card"><span className={`payment-mark ${payment.status === 'paid' ? 'payment-success' : payment.status === 'cancelled' ? 'payment-cancelled' : ''}`}>{payment.status === 'paid' ? <Check size={30} /> : <ShieldCheck size={30} />}</span><span className="eyebrow">پرداخت MALEKI</span><h1>{payment.status === 'paid' ? 'پرداخت با موفقیت انجام شد' : payment.status === 'cancelled' ? 'پرداخت لغو شد' : 'تکمیل خرید'}</h1><p>شماره سفارش #{fa(payment.order_id)}</p><div className="payment-amount"><small>مبلغ قابل پرداخت</small><strong>{money(payment.amount)}</strong></div>
    {payment.ref_id && <div className="payment-reference"><span>شناسه پرداخت آزمایشی</span><b>{payment.ref_id}</b></div>}
    {!done ? <><button className="button button-red button-wide" disabled={busy} onClick={() => confirm('pay')}>{busy ? 'در حال بررسی…' : 'پرداخت و ثبت سفارش'} <ArrowLeft size={16} /></button><button className="payment-cancel" disabled={busy} onClick={() => confirm('cancel')}>انصراف از پرداخت</button></> : <button className="button button-dark button-wide" onClick={() => onNavigate('/account/orders')}>مشاهده سفارش‌ها <ArrowLeft size={16} /></button>}
    <small className="payment-note"><ShieldCheck size={14} /> پرداخت آزمایشی — برای اتصال درگاه واقعی نیاز به تنظیم ارائه‌دهنده دارید.</small></div></div>;
}

function CheckoutPage({ items, session, setQuantity, remove, onClear, onNavigate, onLogin, toast }) {
  const [addresses, setAddresses] = useState([]); const [addressesLoading, setAddressesLoading] = useState(Boolean(session));
  const [addressID, setAddressID] = useState(''); const [shippingMethods, setShippingMethods] = useState([]); const [shippingID, setShippingID] = useState('');
  const [couponOpen, setCouponOpen] = useState(false); const [couponCode, setCouponCode] = useState(''); const [couponResult, setCouponResult] = useState(null);
  const [couponBusy, setCouponBusy] = useState(false); const [couponError, setCouponError] = useState(''); const [busy, setBusy] = useState(false);
  const total = items.reduce((sum, item) => sum + activePrice(item) * item.quantity, 0);
  const selectedAddress = addresses.find((item) => String(item.id) === String(addressID));
  const shippingMethod = shippingMethods.find((item) => String(item.id) === String(shippingID));
  const couponDiscount = couponResult?.subtotal === total ? Number(couponResult.discount_amount || 0) : 0;
  const payable = Math.max(0, total - couponDiscount + Number(shippingMethod?.price || 0));

  useEffect(() => {
    if (!session) { setAddressesLoading(false); return; }
    let current = true; setAddressesLoading(true);
    api('/addresses').then((list) => {
      if (!current) return;
      setAddresses(list);
      setAddressID((old) => old && list.some((item) => String(item.id) === old) ? old : String(list.find((item) => item.is_default)?.id || list[0]?.id || ''));
    }).catch((failure) => { if (current) toast(failure.message, 'error'); }).finally(() => { if (current) setAddressesLoading(false); });
    return () => { current = false; };
  }, [session, toast]);

  useEffect(() => {
    if (!selectedAddress) { setShippingMethods([]); setShippingID(''); return undefined; }
    let current = true; setShippingMethods([]); setShippingID('');
    api(`/shipping-methods?province=${encodeURIComponent(selectedAddress.province)}&city=${encodeURIComponent(selectedAddress.city)}`, { auth: false })
      .then((methods) => { if (current) { setShippingMethods(methods); setShippingID(String(methods[0]?.id || '')); } })
      .catch((failure) => { if (current) { setShippingMethods([]); setShippingID(''); toast(failure.message, 'error'); } });
    return () => { current = false; };
  }, [selectedAddress?.id, selectedAddress?.province, selectedAddress?.city, toast]);

  useEffect(() => { setCouponResult(null); setCouponError(''); }, [total]);

  const applyCoupon = async () => {
    if (!couponCode.trim()) { setCouponError('کد تخفیف را وارد کن.'); return; }
    setCouponBusy(true); setCouponError(''); setCouponResult(null);
    try {
      const result = await api('/coupons/validate', { method: 'POST', body: { code: couponCode.trim(), subtotal: total } });
      setCouponResult({ ...result, subtotal: total }); toast('کد تخفیف اعمال شد.');
    } catch (failure) { setCouponError(failure.message); }
    finally { setCouponBusy(false); }
  };

  const checkout = async () => {
    if (!session) return;
    if (!items.length) { toast('سبد خرید خالی است.', 'error'); return; }
    if (!selectedAddress) { toast('نشانی تحویل را انتخاب کن.', 'error'); return; }
    if (!shippingMethod) { toast('برای نشانی انتخاب‌شده روش ارسال در دسترس نیست.', 'error'); return; }
    setBusy(true);
    try {
      const order = await api('/orders', { method: 'POST', body: { address_id: Number(addressID), shipping_method_id: Number(shippingID), coupon_code: couponDiscount > 0 ? couponCode.trim() : '', items: items.map((item) => ({ product_id: item.product_id, variant_id: item.variant_id || undefined, quantity: item.quantity })) } });
      const payment = await api(`/orders/${order.id}/pay`, { method: 'POST' });
      onClear(); onNavigate(`/pay/${payment.authority}`);
    } catch (failure) { toast(failure.message, 'error'); }
    finally { setBusy(false); }
  };

  if (!session) return <LoginPage onLogin={onLogin} />;
  if (!items.length) return <section className="page-width checkout-page"><div className="checkout-empty"><EmptyState icon={ShoppingBag} title="سبد خرید خالی است" body="برای دیدن مدل‌ها به فروشگاه برگرد." action={<button className="button button-red" onClick={() => onNavigate('/')}>رفتن به فروشگاه <ArrowLeft size={16} /></button>} /></div></section>;

  return <section className="page-width checkout-page">
    <div className="checkout-heading"><div><span className="eyebrow">خرید MALEKI</span><h1>بازبینی و تسویه‌حساب</h1><p>محصول‌ها، نشانی، روش ارسال و مبلغ نهایی را همین‌جا بررسی کن.</p></div><a className="text-link" href={routeHref('/shop')}>ادامهٔ خرید <ArrowLeft size={15} /></a></div>
    <div className="checkout-grid"><div className="checkout-main">
      <section className="checkout-card"><div className="checkout-card-title"><span className="checkout-step">۱</span><div><h2>محصول‌های سبد</h2><small>{fa(items.reduce((sum, item) => sum + item.quantity, 0))} عدد</small></div></div>
        <div className="checkout-items">{items.map((item) => <article className="checkout-item" key={item.key}><a className="checkout-item-image" href={routeHref(`/products/${encodeURIComponent(item.slug || '')}`)}>{item.image_url ? <img src={item.image_url} alt="" /> : <ProductPlaceholder />}</a><div className="checkout-item-info"><a href={routeHref(`/products/${encodeURIComponent(item.slug || '')}`)}><b>{item.name}</b></a>{item.variant_name && <small>{item.variant_label || 'گزینه'}: {item.variant_name}</small>}<div className="checkout-item-price">{item.discount_price > 0 && <del>{money(item.price)}</del>}<strong>{money(activePrice(item))}</strong></div></div><div className="checkout-item-actions"><div className="quantity-stepper"><button onClick={() => setQuantity(item.key, item.quantity - 1)} aria-label="کم کردن تعداد"><Minus size={13} /></button><span>{fa(item.quantity)}</span><button onClick={() => setQuantity(item.key, item.quantity + 1)} disabled={item.quantity >= item.stock} aria-label="زیاد کردن تعداد"><Plus size={13} /></button></div><button className="remove-line" onClick={() => remove(item.key)} aria-label="حذف محصول"><Trash2 size={15} /></button></div><b className="checkout-line-total">{money(activePrice(item) * item.quantity)}</b></article>)}</div>
      </section>

      <section className="checkout-card"><div className="checkout-card-title"><span className="checkout-step">۲</span><div><h2>نشانی تحویل</h2><small>سفارش به این نشانی ارسال می‌شود</small></div><button className="text-link" onClick={() => onNavigate('/account/addresses')}>مدیریت نشانی‌ها <ArrowUpLeft size={14} /></button></div>
        {addressesLoading ? <LoadingState /> : addresses.length ? <><label className="field checkout-address-select"><span>انتخاب نشانی</span><select value={addressID} onChange={(event) => setAddressID(event.target.value)}>{addresses.map((address) => <option value={address.id} key={address.id}>{address.receiver_name} · {address.province}، {address.city}{address.is_default ? ' · پیش‌فرض' : ''}</option>)}</select></label>
          {selectedAddress && <div className="checkout-address-summary"><MapPin size={17} /><div><b>{selectedAddress.receiver_name} <span dir="ltr">{selectedAddress.receiver_phone}</span></b><p>{[selectedAddress.province, selectedAddress.city, selectedAddress.address_line, selectedAddress.postal_code].filter(Boolean).join('، ')}</p></div></div>}</>
          : <div className="checkout-no-address"><p>برای ارسال سفارش، ابتدا یک نشانی ثبت کن.</p><button className="button button-outline button-small" onClick={() => onNavigate('/account/addresses')}>افزودن نشانی <ArrowLeft size={14} /></button></div>}
      </section>

      <section className="checkout-card"><div className="checkout-card-title"><span className="checkout-step">۳</span><div><h2>روش ارسال</h2><small>{selectedAddress ? `${selectedAddress.province}، ${selectedAddress.city}` : 'ابتدا نشانی را انتخاب کن'}</small></div></div>
        {!selectedAddress ? <p className="muted">پس از انتخاب نشانی، روش‌های قابل‌استفاده نشان داده می‌شوند.</p> : shippingMethods.length ? <div className="shipping-choice-list">{shippingMethods.map((method) => <label className={String(method.id) === String(shippingID) ? 'shipping-choice shipping-choice-active' : 'shipping-choice'} key={method.id}><input type="radio" name="shipping-method" value={method.id} checked={String(method.id) === String(shippingID)} onChange={(event) => setShippingID(event.target.value)} /><span><b>{method.name}</b>{method.description && <small>{method.description}</small>}</span><strong>{Number(method.price) === 0 ? 'رایگان' : money(method.price)}</strong></label>)}</div> : <div className="checkout-no-address"><p>برای این نشانی روش ارسالی تعریف نشده است.</p><small>نشانی دیگری انتخاب کن یا با پشتیبانی تماس بگیر.</small></div>}
      </section>
    </div>

    <aside className="checkout-sidebar"><section className="checkout-card checkout-total-card"><span className="eyebrow">خلاصهٔ سفارش</span><h2>پرداخت نهایی</h2>
      <div className="checkout-summary-lines"><div className="summary-row"><span>جمع محصولات</span><b>{money(total)}</b></div><div className="summary-row"><span>هزینهٔ ارسال</span><b>{shippingMethod ? (Number(shippingMethod.price) === 0 ? 'رایگان' : money(shippingMethod.price)) : 'پس از انتخاب روش'}</b></div>{couponDiscount > 0 && <div className="summary-row"><span>تخفیف کد</span><b className="coupon-discount">−{money(couponDiscount)}</b></div>}</div>
      <div className="checkout-coupon"><button type="button" className="coupon-toggle" aria-expanded={couponOpen} onClick={() => setCouponOpen((open) => !open)}><span>{couponDiscount ? `کد ${couponCode.trim()} اعمال شد` : 'کد تخفیف دارید؟'}</span><ChevronDown size={16} className={couponOpen ? 'coupon-chevron coupon-chevron-open' : 'coupon-chevron'} /></button>
        {couponOpen && <div className="coupon-entry checkout-coupon-entry"><label className="field"><span>کد تخفیف</span><input dir="ltr" autoComplete="off" value={couponCode} onChange={(event) => { setCouponCode(event.target.value); setCouponResult(null); setCouponError(''); }} placeholder="کد را وارد کن" /></label><button type="button" className="button button-outline button-small" onClick={applyCoupon} disabled={couponBusy}>{couponBusy ? 'بررسی…' : 'اعمال'}</button>{couponError && <small className="coupon-message coupon-error">{couponError}</small>}{couponDiscount > 0 && <small className="coupon-message">{money(couponDiscount)} تخفیف اعمال شد.</small>}</div>}
      </div>
      <div className="checkout-grand-total"><span>مبلغ قابل پرداخت</span><strong>{money(payable)}</strong></div><button className="button button-red button-wide" onClick={checkout} disabled={busy || addressesLoading || !selectedAddress || !shippingMethod}>{busy ? 'در حال ثبت سفارش…' : 'ثبت سفارش و پرداخت'} <ArrowLeft size={17} /></button>
      <p className="checkout-payment-note"><ShieldCheck size={15} /> مبلغ نهایی در سرور دوباره محاسبه می‌شود. درگاه فعلی آزمایشی است.</p>
    </section></aside></div>
  </section>;
}

function LoadingState() { return <div className="panel-loading"><span className="spinner" /><span>در حال بارگذاری…</span></div>; }

function CampaignSlider() {
  const [slides, setSlides] = useState([]); const [active, setActive] = useState(0);
  const bannerDrag = useRef(null);
  const suppressBannerClick = useRef(false);
  useEffect(() => { api('/banners', { auth: false }).then(setSlides).catch(() => setSlides([])); }, []);
  useEffect(() => { if (slides.length < 2) return undefined; const timer = window.setInterval(() => setActive((value) => (value + 1) % slides.length), 6500); return () => window.clearInterval(timer); }, [slides.length, active]);
  if (!slides.length) return <section className="campaign-slider campaign-fallback page-width" aria-label="ویترین گالری ملکی">
    <div className="campaign-fallback-copy"><span className="eyebrow">MALEKI · JEWELRY GALLERY</span><h1>درخششِ<br /><em>ماندگار</em></h1><p>زیورآلاتی برای لحظه‌هایی که ارزش به‌یادماندن دارند.</p><a className="button button-green" href={routeHref('/shop')}>ورود به گالری <ArrowLeft size={16} /></a></div>
    <div className="campaign-fallback-art" aria-hidden="true"><span className="campaign-gold-orbit campaign-gold-orbit-one" /><span className="campaign-gold-orbit campaign-gold-orbit-two" /><span className="campaign-gold-spark campaign-gold-spark-one">✦</span><span className="campaign-gold-spark campaign-gold-spark-two">✧</span><Gem size={174} strokeWidth={.65} /><small>MALEKI · ۱۴۰۵</small></div>
  </section>;
  const activeIndex = active % slides.length;
  const onPointerDown = (event) => {
    if (slides.length < 2 || event.target.closest('.campaign-dots') || event.button !== 0) return;
    bannerDrag.current = { startX: event.clientX, deltaX: 0, moved: false };
  };
  const onPointerMove = (event) => {
    if (!bannerDrag.current) return;
    const delta = event.clientX - bannerDrag.current.startX;
    bannerDrag.current.deltaX = delta;
    if (!bannerDrag.current.moved && Math.abs(delta) < 8) return;
    if (!bannerDrag.current.moved) event.currentTarget.setPointerCapture(event.pointerId);
    bannerDrag.current.moved = true;
    event.preventDefault();
  };
  const onPointerUp = () => {
    if (!bannerDrag.current) return;
    if (bannerDrag.current.moved) {
      const direction = bannerDrag.current.deltaX < 0 ? 1 : -1;
      setActive((index) => (index + direction + slides.length) % slides.length);
      suppressBannerClick.current = true;
      window.setTimeout(() => { suppressBannerClick.current = false; }, 0);
    }
    bannerDrag.current = null;
  };
  const onClickCapture = (event) => {
    if (!suppressBannerClick.current) return;
    event.preventDefault();
    event.stopPropagation();
    suppressBannerClick.current = false;
  };
  return <section className="campaign-slider page-width" aria-label="بنرهای فروشگاه"
    onPointerDown={onPointerDown} onPointerMove={onPointerMove} onPointerUp={onPointerUp} onPointerCancel={onPointerUp}
    onClickCapture={onClickCapture}>
    {slides.map((slide, index) => <a href={safeCampaignLink(slide.link_url)} className={`campaign-slide${index === activeIndex ? ' campaign-slide-active' : ''}`}
      key={slide.id} aria-hidden={index !== activeIndex} tabIndex={index === activeIndex ? 0 : -1} onDragStart={(event) => event.preventDefault()}>
      <picture><source media="(max-width: 640px)" srcSet={slide.mobile_image_url} /><img src={slide.desktop_image_url} alt={slide.title || 'بنر فروشگاه MALEKI'} /></picture>
      {(slide.title || slide.subtitle || slide.button_label) && <span className="campaign-copy"><b>{slide.title}</b>{slide.subtitle && <small>{slide.subtitle}</small>}{slide.button_label && <em>{slide.button_label}<ArrowLeft size={15} /></em>}</span>}
    </a>)}
    {slides.length > 1 && <div className="campaign-dots">{slides.map((item, index) => <button type="button" aria-label={`بنر ${fa(index + 1)}`} className={index === activeIndex ? 'active' : ''} key={item.id} onClick={() => setActive(index)} />)}</div>}
  </section>;
}

function safeCampaignLink(raw) {
  const value = String(raw || '').trim();
  if (!value || value.startsWith('//')) return '#catalog';
  if (value.startsWith('/') || value.startsWith('#')) return value;
  try {
    const target = new URL(value, window.location.origin);
    return target.protocol === 'https:' ? value : '#catalog';
  } catch { return '#catalog'; }
}

function TrackingPage({ code }) {
  const [order, setOrder] = useState(null); const [error, setError] = useState('');
  useEffect(() => { api(`/tracking/${encodeURIComponent(code)}`, { auth: false }).then(setOrder).catch((failure) => setError(failure.message)); }, [code]);
  if (error) return <div className="page-width section-space"><ErrorPanel message={error} /></div>;
  if (!order) return <div className="page-width loading-view"><span className="spinner" />در حال بررسی کد پیگیری…</div>;
  return <section className="page-width tracking-page"><span className="eyebrow">پیگیری سفارش</span><h1>وضعیت سفارش شما</h1><div className="tracking-card"><div><small>کد پیگیری</small><b dir="ltr">{order.tracking_code}</b></div><div><small>وضعیت</small><strong className={`status-pill status-${order.status}`}>{ORDER_LABELS[order.status] || order.status}</strong></div><div><small>روش ارسال</small><b>{order.shipping_method_name}</b></div><div><small>آخرین به‌روزرسانی</small><b>{dateTime(order.updated_at)}</b></div></div><p className="muted">وضعیت سفارش پس از هر مرحلهٔ پردازش در همین صفحه به‌روز می‌شود.</p><a className="button button-outline" href={routeHref('/account/orders')}>بازگشت به سفارش‌ها <ArrowLeft size={15} /></a></section>;
}

function AdminBanners({ toast }) {
  const blank = { title: '', subtitle: '', desktop_image_url: '', mobile_image_url: '', link_url: '', button_label: '', sort_order: 0, is_active: true };
  const [items, setItems] = useState([]); const [form, setForm] = useState(blank); const [editing, setEditing] = useState(null); const [busy, setBusy] = useState(false);
  const load = () => api('/admin/banners').then(setItems).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, []);
  const set = (key, value) => setForm((item) => ({ ...item, [key]: value }));
  const upload = async (key, file) => { if (!file) return; const data = new FormData(); data.append('file', file); try { const result = await api('/admin/uploads', { method: 'POST', body: data }); set(key, result.url); } catch (failure) { toast(failure.message, 'error'); } };
  const submit = async (event) => { event.preventDefault(); setBusy(true); try { await api(editing ? `/admin/banners/${editing}` : '/admin/banners', { method: editing ? 'PUT' : 'POST', body: { ...form, sort_order: Number(form.sort_order) || 0 } }); setForm(blank); setEditing(null); await load(); toast('بنر ذخیره شد.'); } catch (failure) { toast(failure.message, 'error'); } finally { setBusy(false); } };
  const remove = async (item) => { if (!confirm(`بنر «${item.title || item.id}» حذف شود؟`)) return; try { await api(`/admin/banners/${item.id}`, { method: 'DELETE' }); await load(); toast('بنر حذف شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className="admin-content admin-two-column"><form className="admin-form-panel" onSubmit={submit}><span className="eyebrow">ویترین صفحهٔ اصلی</span><h2>{editing ? 'ویرایش بنر' : 'افزودن بنر یا اسلاید'}</h2>
    <label className="field"><span>عنوان</span><input value={form.title} onChange={(event) => set('title', event.target.value)} /></label><label className="field"><span>زیرعنوان</span><textarea rows={2} value={form.subtitle} onChange={(event) => set('subtitle', event.target.value)} /></label>
    {['desktop_image_url', 'mobile_image_url'].map((key) => <label className="field banner-upload-field" key={key}><span>{key === 'desktop_image_url' ? 'تصویر دسکتاپ' : 'تصویر موبایل'}</span>{form[key] && <img src={form[key]} alt="پیش‌نمایش بنر" />}<input type="file" accept="image/jpeg,image/png,image/webp" required={!form[key]} onChange={(event) => { upload(key, event.target.files?.[0]); event.target.value = ''; }} /><small>تصویر متناسب با دستگاه را جداگانه بارگذاری کن.</small></label>)}
    <label className="field"><span>لینک مقصد</span><input dir="ltr" placeholder="/#catalog یا نشانی داخلی" value={form.link_url} onChange={(event) => set('link_url', event.target.value)} /><small>با کلیک روی بنر، مشتری به این نشانی می‌رود؛ خالی بماند، به فهرست محصولات می‌رود.</small></label><div className="form-grid"><label className="field"><span>متن دکمه (اختیاری)</span><input value={form.button_label} onChange={(event) => set('button_label', event.target.value)} /><small>متن کوتاهی که روی بنر کنار فلش دیده می‌شود؛ مثلاً «دیدن مدل‌ها».</small></label><label className="field"><span>اولویت نمایش</span><input type="number" min="0" value={form.sort_order} onChange={(event) => set('sort_order', event.target.value)} /><small>عدد کمتر زودتر نمایش داده می‌شود؛ ۰ یعنی اولویت اول.</small></label></div>
    <label className="active-checkbox"><input type="checkbox" checked={Boolean(form.is_active)} onChange={(event) => set('is_active', event.target.checked)} /><span><b>نمایش در صفحهٔ اصلی</b></span></label><button className="button button-red" disabled={busy}>ذخیره بنر <ArrowLeft size={15} /></button>{editing && <button type="button" className="text-link" onClick={() => { setEditing(null); setForm(blank); }}>انصراف از ویرایش</button>}</form>
    <div className="admin-list-panel"><span className="eyebrow">مدیریت کمپین</span><h2>بنرها و اسلایدها</h2><div className="simple-admin-list">{items.map((item) => <div className="simple-admin-row banner-admin-row" key={item.id}><img src={item.desktop_image_url} alt="" /><span><b>{item.title || 'بدون عنوان'}</b><small>{item.is_active ? 'فعال' : 'پنهان'} · ترتیب {fa(item.sort_order)}</small></span><div className="row-actions"><button onClick={() => { setEditing(item.id); setForm({ ...item }); }}><PenIcon /></button><button onClick={() => remove(item)}><Trash2 size={15} /></button></div></div>)}{!items.length && <p className="muted">هنوز بنری ثبت نشده است.</p>}</div></div></div>;
}

function AdminShipping({ toast }) {
  const blank = { name: '', description: '', province: '', city: '', price: '', is_active: true, sort_order: 0 };
  const [items, setItems] = useState([]); const [form, setForm] = useState(blank); const [editing, setEditing] = useState(null);
  const load = () => api('/admin/shipping-methods').then(setItems).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, []);
  const set = (key, value) => setForm((item) => ({ ...item, [key]: value }));
  const submit = async (event) => { event.preventDefault(); const price = parseWholeNumber(form.price); if (price === null) { toast('هزینه را با عدد صحیح وارد کن؛ مثلاً ۱۵۰٬۰۰۰.', 'error'); return; } try { await api(editing ? `/admin/shipping-methods/${editing}` : '/admin/shipping-methods', { method: editing ? 'PUT' : 'POST', body: { ...form, price, sort_order: Number(form.sort_order) || 0 } }); setForm(blank); setEditing(null); await load(); toast('روش ارسال ذخیره شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  const remove = async (item) => { if (!confirm(`روش «${item.name}» حذف شود؟`)) return; try { await api(`/admin/shipping-methods/${item.id}`, { method: 'DELETE' }); await load(); toast('روش ارسال حذف شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className="admin-content admin-two-column"><form className="admin-form-panel" onSubmit={submit}><span className="eyebrow">تحویل سفارش</span><h2>{editing ? 'ویرایش روش ارسال' : 'افزودن روش ارسال'}</h2><label className="field"><span>عنوان روش</span><input required value={form.name} onChange={(event) => set('name', event.target.value)} placeholder="پست، تیپاکس یا پیک" /></label><label className="field"><span>توضیحات</span><input value={form.description} onChange={(event) => set('description', event.target.value)} /></label><label className="field"><span>استان مقصد</span><select value={form.province} onChange={(event) => setForm((current) => ({ ...current, province: event.target.value, city: '' }))}><option value="">همهٔ استان‌ها</option>{IRAN_PROVINCES.map((province) => <option key={province}>{province}</option>)}</select></label><label className="field"><span>شهر مقصد (اختیاری)</span><input value={form.city} onChange={(event) => set('city', event.target.value)} placeholder="مثلاً تهران؛ خالی برای کل استان" /></label><div className="form-grid"><label className="field"><span>هزینهٔ ارسال (تومان)</span><input type="text" inputMode="numeric" dir="ltr" autoComplete="off" required value={form.price} onChange={(event) => set('price', event.target.value)} placeholder="مثلاً ۱۵۰٬۰۰۰" /><small>این مبلغ در پرداخت به جمع سفارش اضافه می‌شود. عدد ۰ فقط برای ارسال رایگان است.</small></label><label className="field"><span>اولویت در فهرست</span><input type="number" min="0" value={form.sort_order} onChange={(event) => set('sort_order', event.target.value)} /><small>عدد کمتر بالاتر نشان داده می‌شود؛ ۰ یعنی اولویت اول.</small></label></div><label className="active-checkbox"><input type="checkbox" checked={Boolean(form.is_active)} onChange={(event) => set('is_active', event.target.checked)} /><span><b>فعال</b></span></label><button className="button button-red">ذخیره روش ارسال <ArrowLeft size={15} /></button>{editing && <button className="button button-outline" type="button" onClick={() => { setEditing(null); setForm(blank); }}>لغو ویرایش</button>}</form>
    <div className="admin-list-panel"><span className="eyebrow">تعرفه‌های فروشگاه</span><h2>روش‌های ارسال</h2><div className="simple-admin-list">{items.map((item) => <div className="simple-admin-row" key={item.id}><span className="simple-row-icon"><Truck size={17} /></span><span><b>{item.name} · {Number(item.price) === 0 ? 'رایگان' : money(item.price)}</b><small>{item.province ? `${item.province}${item.city ? `، ${item.city}` : ''}` : 'سراسری'} · {item.is_active ? 'فعال' : 'غیرفعال'} · هزینه: {money(item.price)}</small></span><div className="row-actions"><button type="button" aria-label={`ویرایش ${item.name}`} onClick={() => { setEditing(item.id); setForm({ ...item, price: String(item.price) }); }}><PenIcon /></button><button type="button" aria-label={`حذف ${item.name}`} onClick={() => remove(item)}><Trash2 size={15} /></button></div></div>)}</div>{!items.length && <p className="muted">هنوز روش ارسالی ثبت نشده است.</p>}</div></div>;
}

function AdminCoupons({ toast }) {
  // An empty minimum is the same as no purchase minimum (0 Toman). Keeping
  // the input blank lets admins clear the value without triggering validation.
  const blank = { code: '', kind: 'percent', value: '', minimum_subtotal: '', maximum_discount: '', usage_limit: '', starts_at: '', ends_at: '', is_active: true };
  const [items, setItems] = useState([]); const [form, setForm] = useState(blank); const [editing, setEditing] = useState(null);
  const load = () => api('/admin/coupons').then(setItems).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, []);
  const set = (key, value) => setForm((item) => ({ ...item, [key]: value }));
  const submit = async (event) => { event.preventDefault(); const value = parseWholeNumber(form.value); const minimum = form.minimum_subtotal.trim() === '' ? 0 : parseWholeNumber(form.minimum_subtotal); const maximum = form.maximum_discount ? parseWholeNumber(form.maximum_discount) : null; if (value === null || minimum === null || (form.maximum_discount && maximum === null)) { toast('مبلغ‌ها را با عدد صحیح وارد کن؛ مثلاً ۱۵۰٬۰۰۰.', 'error'); return; } const body = { ...form, code: form.code.trim().toUpperCase(), value, minimum_subtotal: minimum, maximum_discount: maximum, usage_limit: form.usage_limit ? Number(form.usage_limit) : null, starts_at: form.starts_at ? new Date(form.starts_at).toISOString() : null, ends_at: form.ends_at ? new Date(form.ends_at).toISOString() : null }; try { await api(editing ? `/admin/coupons/${editing}` : '/admin/coupons', { method: editing ? 'PUT' : 'POST', body }); setForm(blank); setEditing(null); await load(); toast('کد تخفیف ذخیره شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  const edit = (item) => { setEditing(item.id); setForm({ ...item, value: String(item.value), minimum_subtotal: item.minimum_subtotal ? String(item.minimum_subtotal) : '', maximum_discount: item.maximum_discount ? String(item.maximum_discount) : '', usage_limit: item.usage_limit ? String(item.usage_limit) : '', starts_at: item.starts_at ? new Date(item.starts_at).toISOString().slice(0,16) : '', ends_at: item.ends_at ? new Date(item.ends_at).toISOString().slice(0,16) : '' }); };
  const remove = async (item) => { if (!confirm(`کد ${item.code} حذف شود؟`)) return; try { await api(`/admin/coupons/${item.id}`, { method: 'DELETE' }); await load(); toast('کد تخفیف حذف شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className="admin-content admin-two-column"><form className="admin-form-panel" onSubmit={submit}><span className="eyebrow">پیشنهادهای فروش</span><h2>{editing ? 'ویرایش کد' : 'ساخت کد تخفیف'}</h2><label className="field"><span>کد</span><input dir="ltr" required value={form.code} onChange={(event) => set('code', event.target.value.toUpperCase())} /></label><div className="form-grid"><label className="field"><span>نوع تخفیف</span><select value={form.kind} onChange={(event) => set('kind', event.target.value)}><option value="percent">درصدی</option><option value="fixed">مبلغ ثابت</option></select></label><label className="field"><span>{form.kind === 'percent' ? 'درصد' : 'مبلغ تخفیف به تومان'}</span><input type="text" inputMode="numeric" dir="ltr" required value={form.value} onChange={(event) => set('value', event.target.value)} placeholder={form.kind === 'percent' ? 'مثلاً ۱۰' : 'مثلاً ۱۵۰٬۰۰۰'} /><small>{form.kind === 'percent' ? 'عدد بین ۱ تا ۱۰۰ وارد کن.' : 'مبلغ ثابت تخفیف؛ به تومان.'}</small></label></div><label className="field"><span>حداقل مبلغ سبد (تومان)</span><input type="text" inputMode="numeric" dir="ltr" value={form.minimum_subtotal} onChange={(event) => set('minimum_subtotal', event.target.value)} placeholder="بدون حداقل" /><small>خالی یا ۰ یعنی این کد حداقل مبلغ خرید ندارد.</small></label>{form.kind === 'percent' && <label className="field"><span>سقف تخفیف (تومان، اختیاری)</span><input type="text" inputMode="numeric" dir="ltr" value={form.maximum_discount} onChange={(event) => set('maximum_discount', event.target.value)} placeholder="بدون سقف" /></label>}<div className="form-grid"><label className="field"><span>تعداد استفاده مجاز</span><input type="number" min="1" placeholder="بدون محدودیت" value={form.usage_limit} onChange={(event) => set('usage_limit', event.target.value)} /></label><label className="field"><span>تاریخ شروع</span><input type="datetime-local" value={form.starts_at || ''} onChange={(event) => set('starts_at', event.target.value)} /></label><label className="field"><span>تاریخ پایان</span><input type="datetime-local" value={form.ends_at || ''} onChange={(event) => set('ends_at', event.target.value)} /></label></div><label className="active-checkbox"><input type="checkbox" checked={Boolean(form.is_active)} onChange={(event) => set('is_active', event.target.checked)} /><span><b>فعال</b></span></label><button className="button button-red">ذخیره کد تخفیف <ArrowLeft size={15} /></button></form>
    <div className="admin-list-panel"><span className="eyebrow">وضعیت مصرف</span><h2>کدهای تخفیف</h2><p className="muted">درصدی: تخفیف به‌صورت درصد از مبلغ محصولات محاسبه می‌شود. مبلغ ثابت: همین مبلغ از جمع محصولات کم می‌شود. حداقل خرید نیز شرط استفاده از کد است.</p><div className="simple-admin-list">{items.map((item) => <div className="simple-admin-row" key={item.id}><span className="simple-row-icon">%</span><span><b dir="ltr">{item.code || 'کد بدون عنوان'} · {item.kind === 'percent' ? `${fa(item.value)}٪ تخفیف` : `${money(item.value)} تخفیف`}</b><small>حداقل خرید: {Number(item.minimum_subtotal) > 0 ? money(item.minimum_subtotal) : 'بدون حداقل'} · استفاده‌شده {fa(item.used_count)}{item.usage_limit ? ` از ${fa(item.usage_limit)}` : ''} · {item.is_active ? 'فعال' : 'غیرفعال'}</small></span><div className="row-actions"><button type="button" aria-label={`ویرایش کد ${item.code}`} onClick={() => edit(item)}><PenIcon /></button><button type="button" aria-label={`حذف کد ${item.code}`} onClick={() => remove(item)}><Trash2 size={15} /></button></div></div>)}</div>{!items.length && <p className="muted">هنوز کد تخفیفی ثبت نشده است.</p>}</div></div>;
}

const ADMIN_TABS = [
  ['/admin', 'پیشخوان', '⌂'], ['/admin/products', 'محصولات', '▦'], ['/admin/categories', 'دسته‌بندی‌ها', '▤'],
  ['/admin/brands', 'برندها', '◇'], ['/admin/orders', 'سفارش‌ها', '▣'], ['/admin/users', 'کاربران', '♙'],
  ['/admin/comments', 'دیدگاه‌ها', '◌'], ['/admin/blog', 'وبلاگ', '✎'], ['/admin/tickets', 'تیکت‌ها', '✉'],
  ['/admin/banners', 'بنر و اسلایدر', '▧'], ['/admin/shipping', 'روش‌های ارسال', '↗'], ['/admin/coupons', 'کدهای تخفیف', '%'],
  ['/admin/analytics', 'آمار و تحلیل', '◷'], ['/admin/site-content', 'محتوای سایت', '¶'],
];

function AdminPage({ route, categories, brands, toast }) {
  const detailTicket = route.path.match(/^\/admin\/tickets\/(\d+)$/);
  const detailOrder = route.path.match(/^\/admin\/orders\/(\d+)$/);
  const productForm = route.path.match(/^\/admin\/products\/(new|edit\/(\d+))$/);
  const brandForm = route.path.match(/^\/admin\/brands\/(new|edit\/(\d+))$/);
  const key = detailTicket ? 'tickets' : route.path.split('/')[2] || 'dashboard';
  const title = key === 'products' && productForm
    ? productForm[2] ? 'ویرایش محصول' : 'افزودن محصول جدید'
    : key === 'brands' && brandForm
      ? brandForm[2] ? 'ویرایش برند' : 'افزودن برند جدید'
      : ADMIN_TABS.find(([href]) => (href === '/admin' ? key === 'dashboard' : route.path.startsWith(href)))?.[1] || 'مدیریت';
  const [productsMenuOpen, setProductsMenuOpen] = useState(route.path.startsWith('/admin/products'));
  const [brandsMenuOpen, setBrandsMenuOpen] = useState(route.path.startsWith('/admin/brands'));
  useEffect(() => { if (route.path.startsWith('/admin/products')) setProductsMenuOpen(true); }, [route.path]);
  useEffect(() => { if (route.path.startsWith('/admin/brands')) setBrandsMenuOpen(true); }, [route.path]);
  let view;
  if (key === 'products') view = <AdminProducts categories={categories} brands={brands} toast={toast} formPage={Boolean(productForm)} editID={productForm?.[2] || ''} />;
  else if (key === 'categories') view = <AdminCategories categories={categories} toast={toast} />;
  else if (key === 'brands') view = <AdminBrands brands={brands} toast={toast} formPage={Boolean(brandForm)} editID={brandForm?.[2] || ''} />;
  else if (key === 'orders') view = detailOrder ? <AdminOrderDetail id={detailOrder[1]} toast={toast} /> : <AdminOrders toast={toast} />;
  else if (key === 'users') view = <AdminUsers toast={toast} />;
  else if (key === 'comments') view = <AdminComments toast={toast} />;
  else if (key === 'blog') view = <AdminBlog toast={toast} />;
  else if (key === 'tickets') view = detailTicket ? <AdminTicketDetail id={detailTicket[1]} toast={toast} /> : <AdminTickets toast={toast} />;
  else if (key === 'banners') view = <AdminBanners toast={toast} />;
  else if (key === 'shipping') view = <AdminShipping toast={toast} />;
  else if (key === 'coupons') view = <AdminCoupons toast={toast} />;
  else if (key === 'analytics') view = <AdminAnalytics toast={toast} />;
  else if (key === 'site-content') view = <AdminSiteContent toast={toast} />;
  else view = <AdminDashboard toast={toast} />;
  return <div className="admin-page page-width"><aside className="admin-sidebar"><div className="admin-brand"><span className="brand-mark"><Gem size={19} /></span><span><b>MALEKI</b><small>میز مدیریت فروشگاه</small></span></div><nav>{ADMIN_TABS.map(([href, label, mark]) => {
    if (href === '/admin/products' || href === '/admin/brands') {
      const productsGroup = href === '/admin/products';
      const groupKey = productsGroup ? 'products' : 'brands';
      const isForm = productsGroup ? productForm : brandForm;
      const isOpen = productsGroup ? productsMenuOpen : brandsMenuOpen;
      const setOpen = productsGroup ? setProductsMenuOpen : setBrandsMenuOpen;
      const basePath = href;
      const listLabel = productsGroup ? 'فهرست محصولات' : 'فهرست برندها';
      const addLabel = productsGroup ? 'افزودن محصول جدید' : 'افزودن برند جدید';
      return <div className={`admin-nav-group ${key === groupKey ? 'admin-nav-group-active' : ''}`} key={href}>
        <div className="admin-nav-parent"><a className={key === groupKey && !isForm ? 'active' : ''} href={routeHref(basePath)}><span className="admin-tab-mark">{mark}</span>{label}</a><button type="button" aria-label={`نمایش زیرمنوی ${label}`} aria-expanded={isOpen} onClick={() => setOpen((open) => !open)}><ChevronDown size={15} /></button></div>
        {isOpen && <div className="admin-nav-submenu"><a className={key === groupKey && !isForm ? 'active' : ''} href={routeHref(basePath)}>{listLabel}</a><a className={isForm && !isForm[2] ? 'active' : ''} href={routeHref(`${basePath}/new`)}>{addLabel}</a></div>}
      </div>;
    }
    return <a className={(href === '/admin' ? key === 'dashboard' : route.path.startsWith(href)) ? 'active' : ''} href={routeHref(href)} key={href}><span className="admin-tab-mark">{mark}</span>{label}{key === 'tickets' && href === '/admin/tickets' ? <ArrowLeft size={14} /> : null}</a>;
  })}</nav><a className="admin-store-link" href={routeHref('/')}>← بازگشت به فروشگاه</a></aside>
    <div className="admin-main"><header className="admin-heading"><div><span className="eyebrow">MALEKI · مدیریت</span><h1>{title}</h1><p>مدیریت فروشگاه، از یک نقطه.</p></div><a href={routeHref('/')} className="button button-outline button-small">مشاهده فروشگاه <ArrowUpLeft size={15} /></a></header>{view}</div></div>;
}

function AdminDashboard({ toast }) {
  const [stats, setStats] = useState(null); const [orders, setOrders] = useState([]);
  useEffect(() => { Promise.all([api('/admin/overview'), api('/admin/orders?page=1&limit=5')]).then(([summary, result]) => { setStats(summary); setOrders(result.data || []); }).catch((failure) => toast(failure.message, 'error')); }, []);
  if (!stats) return <LoadingState />;
  return <div className="admin-content"><div className="admin-welcome"><div><span className="eyebrow">خلاصه امروز</span><h2>سلام مدیر، آماده‌ی حرکتیم؟</h2><p>تصویر روشن فروشگاه و سفارش‌هایی که منتظر رسیدگی‌اند.</p></div><span className="admin-date">{shortDate(new Date().toISOString())}</span></div>
    <div className="admin-stats">{[['فروش ثبت‌شده', money(stats.paid_revenue), 'مجموع سفارش‌های پرداخت‌شده', '↗'], ['سفارش‌ها', fa(stats.orders), `${fa(stats.pending_orders)} سفارش نیازمند اقدام`, '▣'], ['محصولات', fa(stats.products), 'محصولات فعال و غیرفعال', '▦'], ['کاربران', fa(stats.users), `${fa(stats.categories)} دسته‌بندی`, '♙']].map(([label, value, hint, icon]) => <article className="admin-stat-card" key={label}><div className="admin-stat-heading"><small>{label}</small><span className="admin-stat-icon" aria-hidden="true">{icon}</span></div><strong>{value}</strong><span>{hint}</span></article>)}</div>
    <section className="admin-panel"><div className="admin-panel-heading"><div><span className="eyebrow">آخرین حرکت‌ها</span><h2>سفارش‌های تازه</h2></div><a className="text-link" href={routeHref('/admin/orders')}>همه سفارش‌ها <ArrowLeft size={15} /></a></div>{orders.length ? <div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>سفارش</th><th>مشتری</th><th>وضعیت</th><th>مبلغ</th><th>تاریخ</th></tr></thead><tbody>{orders.map((order) => <tr key={order.id}><td>#{fa(order.id)}</td><td>{order.customer_name || order.customer_phone}</td><td><span className={`status-pill status-${order.status}`}>{ORDER_LABELS[order.status]}</span></td><td>{money(order.total_amount)}</td><td>{shortDate(order.created_at)}</td></tr>)}</tbody></table></div> : <p className="muted">هنوز سفارشی ثبت نشده است.</p>}</section>
    <div className="admin-quick-links">{ADMIN_TABS.slice(1, 5).map(([href, label]) => <a href={routeHref(href)} key={href}>{label}<ArrowUpLeft size={15} /></a>)}</div></div>;
}

const emptyProduct = { name: '', slug: '', description: '', image_urls: [], attributes: {}, variant_label: '', variants: [], category_id: '', brand_id: '', price: '', discount_price: '', stock: '', is_active: true, is_popular: false };

function AdminProducts({ categories, brands, toast, formPage, editID }) {
  const [products, setProducts] = useState([]); const [form, setForm] = useState(emptyProduct); const [editing, setEditing] = useState(null); const [search, setSearch] = useState(''); const [stockFilter, setStockFilter] = useState(''); const [busy, setBusy] = useState(false);
  const load = () => api(`/admin/products?page=1&limit=100${search ? `&q=${encodeURIComponent(search)}` : ''}`).then((result) => setProducts(result.data || [])).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, [search]);
  const editProduct = (product) => { setEditing(product.id); setForm({ ...emptyProduct, ...product, category_id: String(product.category_id || ''), brand_id: String(product.brand_id || ''), price: product.price, discount_price: product.discount_price || '', stock: product.stock, image_urls: product.image_urls?.length ? product.image_urls : product.image_url ? [product.image_url] : [], variants: product.variants || [] }); go(`/admin/products/edit/${product.id}`); };
  useEffect(() => {
    if (!formPage) { setEditing(null); setForm(emptyProduct); return; }
    if (!editID) { setEditing(null); setForm(emptyProduct); return; }
    const product = products.find((item) => String(item.id) === String(editID));
    if (product) {
      setEditing(product.id);
      setForm({ ...emptyProduct, ...product, category_id: String(product.category_id || ''), brand_id: String(product.brand_id || ''), price: product.price, discount_price: product.discount_price || '', stock: product.stock, image_urls: product.image_urls?.length ? product.image_urls : product.image_url ? [product.image_url] : [], variants: product.variants || [] });
    }
  }, [formPage, editID, products]);
  const set = (key, value) => setForm((item) => ({ ...item, [key]: value }));
  const setVariant = (index, key, value) => setForm((item) => ({ ...item, variants: item.variants.map((variant, i) => i === index ? { ...variant, [key]: value } : variant) }));
  const setAttribute = (index, field, value) => setForm((item) => {
    const entries = Object.entries(item.attributes || {});
    entries[index] = field === 'key' ? [value, entries[index][1]] : [entries[index][0], value];
    return { ...item, attributes: Object.fromEntries(entries) };
  });
  const addAttribute = () => {
    const attributes = form.attributes || {};
    let index = 1;
    while (Object.hasOwn(attributes, `ویژگی ${index}`)) index += 1;
    set('attributes', { ...attributes, [`ویژگی ${index}`]: '' });
  };
  const removeAttribute = (index) => setForm((item) => ({ ...item, attributes: Object.fromEntries(Object.entries(item.attributes || {}).filter((_, entryIndex) => entryIndex !== index)) }));
  const visibleProducts = products.filter((product) => !stockFilter || (stockFilter === 'available' ? product.stock > 0 : product.stock <= 0));
  const upload = async (files) => { for (const file of files) { const data = new FormData(); data.append('file', file); try { const result = await api('/admin/uploads', { method: 'POST', body: data }); setForm((item) => ({ ...item, image_urls: [...item.image_urls, result.url].slice(0, 8) })); } catch (failure) { toast(failure.message, 'error'); } } };
  const uploadVariantImage = async (index, file) => { if (!file) return; const data = new FormData(); data.append('file', file); try { const result = await api('/admin/uploads', { method: 'POST', body: data }); setVariant(index, 'image_url', result.url); toast('عکس این مقدار بارگذاری شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  const submit = async (event) => {
    event.preventDefault(); setBusy(true);
    const variants = form.variants.map((item) => ({ ...item, price: Number(item.price), discount_price: Number(item.discount_price || 0), stock: Number(item.stock) }));
    const attributes = Object.fromEntries(Object.entries(form.attributes || {}).map(([key, value]) => [key.trim(), String(value || '').trim()]).filter(([key, value]) => key && value));
    const body = { name: form.name.trim(), slug: form.slug.trim(), description: form.description.trim(), image_url: form.image_urls[0] || '', image_urls: form.image_urls, attributes, variant_label: form.variant_label, variants, is_popular: Boolean(form.is_popular),
      category_id: Number(form.category_id), brand_id: Number(form.brand_id), price: variants.length ? 0 : Number(form.price), discount_price: variants.length ? 0 : Number(form.discount_price || 0), stock: variants.length ? 0 : Number(form.stock), is_active: Boolean(form.is_active) };
    try { await api(editing ? `/products/${editing}` : '/products', { method: editing ? 'PUT' : 'POST', body }); setForm(emptyProduct); setEditing(null); await load(); toast(editing ? 'محصول ویرایش شد.' : 'محصول تازه ثبت شد.'); go('/admin/products'); }
    catch (failure) { toast(failure.message, 'error'); } finally { setBusy(false); }
  };
  const toggleActive = async (product) => { try { await api(`/admin/products/${product.id}/visibility`, { method: 'PATCH', body: { is_active: !product.is_active } }); await load(); toast(product.is_active ? 'محصول از فروشگاه پنهان شد.' : 'محصول در فروشگاه فعال شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  const remove = async (product) => { if (!confirm(`محصول «${product.name}» حذف شود؟`)) return; try { await api(`/products/${product.id}`, { method: 'DELETE' }); await load(); toast('محصول حذف شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className={`admin-content admin-products-page ${formPage ? 'admin-product-form-page' : ''}`}>{!formPage && <div className="admin-list-panel"><div className="admin-panel-heading"><div><span className="eyebrow">کاتالوگ فروشگاه</span><h2>محصولات <small>{fa(visibleProducts.length)}</small></h2></div><div className="admin-product-tools"><label className="admin-search"><Search size={16} /><input placeholder="جستجوی محصول" value={search} onChange={(event) => setSearch(event.target.value)} /></label><label className="status-filter"><select aria-label="فیلتر موجودی محصولات" value={stockFilter} onChange={(event) => setStockFilter(event.target.value)}><option value="">همهٔ محصولات</option><option value="available">موجود</option><option value="unavailable">ناموجود</option></select></label><a className="button button-red button-small" href={routeHref('/admin/products/new')}><Plus size={14} /> محصول جدید</a></div></div>
    <div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>محصول</th><th>دسته / برند</th><th>قیمت</th><th>موجودی</th><th>نمایش</th><th></th></tr></thead><tbody>{visibleProducts.map((product) => <tr key={product.id}><td><div className="admin-product-cell">{product.image_url ? <img src={product.image_url} alt="" /> : <span><Gem size={18} /></span>}<div><b>{product.name}</b><small>{product.slug}</small></div></div></td><td>{product.category_name || '—'}<small>{product.brand_name}</small></td><td><PriceDisplay source={product} /></td><td>{fa(product.stock)}</td><td><button className={`visibility-switch ${product.is_active ? 'switch-on' : ''}`} onClick={() => toggleActive(product)}><span />{product.is_active ? 'فعال' : 'پنهان'}</button></td><td><div className="row-actions"><button onClick={() => editProduct(product)} aria-label="ویرایش"><PenIcon /></button><button onClick={() => remove(product)} aria-label="حذف"><Trash2 size={15} /></button></div></td></tr>)}</tbody></table>{!visibleProducts.length && <p className="muted table-empty">محصولی برای این فیلتر موجود نیست.</p>}</div></div>}
    {formPage && <form className="admin-form-panel" onSubmit={submit}><div className="admin-panel-heading"><div><span className="eyebrow">مدیریت کالا</span><h2>{editing ? 'ویرایش محصول' : 'محصول جدید'}</h2></div><div className="admin-form-heading-actions"><a className="text-link" href={routeHref('/admin/products')}>بازگشت به فهرست <ArrowRight size={14} /></a>{editing && <button className="icon-button" type="button" onClick={() => go('/admin/products')} aria-label="لغو ویرایش"><X size={18} /></button>}</div></div>
      <div className="form-stack"><div className="form-grid"><label className="field field-wide"><span>نام محصول</span><input value={form.name} onChange={(event) => { const name = event.target.value; setForm((item) => ({ ...item, name, slug: item.slug ? item.slug : createSlug(name) })); }} required /></label><label className="field"><span>اسلاگ</span><input dir="ltr" value={form.slug} onChange={(event) => set('slug', createSlug(event.target.value))} /></label>
        <label className="field"><span>دسته‌بندی</span><select value={form.category_id} onChange={(event) => set('category_id', event.target.value)} required><option value="">انتخاب دسته</option>{categories.map((item) => <option key={item.id} value={item.id}>{item.parent_id ? `— ${item.name}` : item.name}</option>)}</select></label>
        <label className="field"><span>برند</span><select value={form.brand_id} onChange={(event) => set('brand_id', event.target.value)} required><option value="">انتخاب برند</option>{brands.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>
        <div className="field field-wide"><RichTextEditor label="توضیحات محصول" value={form.description} onChange={(value) => set("description", value)} placeholder="معرفی، جنس، کاربرد و جزئیات محصول را بنویس…" minHeight={180} /></div></div>
        <section className="admin-product-attributes"><div><h3>ویژگی‌های کوتاه محصول</h3><small>تا چهار ویژگی؛ مثل جنس، رنگ یا نوع نگین. در کنار عکس محصول نمایش داده می‌شوند.</small></div>
          {Object.entries(form.attributes || {}).map(([key, value], index) => <div className="admin-product-attribute-row" key={index}><label className="field"><span>عنوان</span><input value={key} onChange={(event) => setAttribute(index, 'key', event.target.value)} placeholder="مثلاً جنس" /></label><label className="field"><span>توضیح کوتاه</span><input value={value} onChange={(event) => setAttribute(index, 'value', event.target.value)} placeholder="مثلاً نقرهٔ ۹۲۵" /></label><button type="button" className="remove-variant" aria-label={`حذف ویژگی ${key || index + 1}`} onClick={() => removeAttribute(index)}><Trash2 size={15} /></button></div>)}
          {Object.keys(form.attributes || {}).length < 4 && <button type="button" className="button button-outline button-small" onClick={addAttribute}><Plus size={14} /> افزودن ویژگی</button>}
        </section>
        <div className="field"><span className="field-label">تصاویر محصول</span><div className="admin-image-grid">{form.image_urls.map((url, index) => <div key={`${url}-${index}`}><img src={url} alt={`تصویر ${index + 1}`} /><button type="button" onClick={() => set('image_urls', form.image_urls.filter((_, i) => i !== index))}><X size={13} /></button>{index === 0 && <small>اصلی</small>}</div>)}<label className="upload-tile"><ImagePlus size={22} /><span>افزودن عکس</span><input type="file" accept="image/jpeg,image/png,image/webp" multiple onChange={(event) => { upload([...event.target.files]); event.target.value = ''; }} /></label></div></div>
        <label className="active-checkbox"><input type="checkbox" checked={Boolean(form.is_popular)} onChange={(event) => set('is_popular', event.target.checked)} /><span><b>نمایش در محبوب‌ترین‌ها</b><small>محصول فعال در اسلایدر صفحهٔ اصلی دیده می‌شود.</small></span></label>
        {!form.variants.length ? <div className="price-editor"><label className="field"><span>قیمت عادی <small>تومان</small></span><input type="number" min="1" value={form.price} onChange={(event) => set('price', event.target.value)} required /></label><label className="field"><span>قیمت تخفیفی <small>اختیاری</small></span><input type="number" min="0" value={form.discount_price} onChange={(event) => set('discount_price', event.target.value)} placeholder="بدون تخفیف" /></label><label className="field"><span>موجودی</span><input type="number" min="0" value={form.stock} onChange={(event) => set('stock', event.target.value)} required /></label>{Number(form.discount_price) > 0 && Number(form.discount_price) < Number(form.price) && <div className="discount-preview"><BadgePercent size={18} /><span>{fa(discountPercent({ price: form.price, discount_price: form.discount_price }))}٪ تخفیف</span><del>{money(form.price)}</del><b>{money(form.discount_price)}</b></div>}</div>
          : <div className="variant-editor"><label className="field"><span>عنوان تنوع‌ها</span><input placeholder="سایز، رنگ یا حجم" value={form.variant_label} onChange={(event) => set('variant_label', event.target.value)} required /></label>{form.variants.map((variant, index) => <div className="variant-admin-row" key={variant.id || index}>
            <label className="field"><span>{form.variant_label || 'مقدار'}</span><input value={variant.name} onChange={(event) => setVariant(index, 'name', event.target.value)} required /></label>
            <label className="field"><span>قیمت عادی</span><input type="number" min="1" value={variant.price} onChange={(event) => setVariant(index, 'price', event.target.value)} required /></label>
            <label className="field"><span>قیمت تخفیفی</span><input type="number" min="0" value={variant.discount_price || ''} onChange={(event) => setVariant(index, 'discount_price', event.target.value)} /></label>
            <label className="field"><span>موجودی</span><input type="number" min="0" value={variant.stock} onChange={(event) => setVariant(index, 'stock', event.target.value)} required /></label>
            <label className="field variant-image-upload"><span>عکس این مقدار</span><input type="file" accept="image/jpeg,image/png,image/webp" onChange={(event) => { uploadVariantImage(index, event.target.files?.[0]); event.target.value = ''; }} />{variant.image_url && <small>✓ عکس ثبت شد</small>}</label>
            <button className="remove-variant" type="button" onClick={() => set('variants', form.variants.filter((_, i) => i !== index))} aria-label="حذف تنوع"><Trash2 size={15} /></button>
            {Number(variant.discount_price) > 0 && Number(variant.discount_price) < Number(variant.price) && <small className="variant-discount-note">تخفیف {fa(discountPercent({ price: variant.price, discount_price: variant.discount_price }))}٪</small>}
          </div>)}<button className="button button-outline button-small" type="button" onClick={() => set('variants', [...form.variants, { id: '', name: '', price: '', discount_price: '', stock: 0, image_url: '' }])}><Plus size={15} />افزودن مقدار</button></div>}
        {!form.variants.length && <button className="text-link variant-add-link" type="button" onClick={() => { set('price', ''); set('stock', 0); set('discount_price', ''); set('variant_label', ''); set('variants', [{ id: '', name: '', price: '', discount_price: '', stock: 0, image_url: '' }]); }}>این محصول گزینه‌های متغیر دارد؟ <Plus size={14} /></button>}
        <label className="active-checkbox"><input type="checkbox" checked={Boolean(form.is_active)} onChange={(event) => set('is_active', event.target.checked)} /><span><b>نمایش محصول در فروشگاه</b><small>با خاموش‌کردن، محصول برای مشتری‌ها پنهان می‌شود.</small></span></label>
        <button className="button button-red button-wide" disabled={busy}>{busy ? 'در حال ذخیره…' : editing ? 'ذخیره تغییرات' : 'ثبت محصول'} <ArrowLeft size={16} /></button></div></form>}</div>;
}

function PriceDisplay({ source }) {
  const variants = source.variants || [];
  const chosen = variants.length ? [...variants].sort((a, b) => activePrice(source, a) - activePrice(source, b))[0] : source;
  const sale = Number(chosen?.discount_price || 0);
  return <div className="table-price">{sale > 0 && <del>{money(chosen.price)}</del>}<b>{variants.length ? `از ${money(activePrice(source, chosen))}` : money(activePrice(source))}</b></div>;
}

function AdminCategories({ categories, toast }) {
  const blank = { name: "", slug: "", description: "", seo_title: "", seo_description: "", home_title: "", home_image_url: "", show_on_home: false, parent_id: "" };
  const [form, setForm] = useState(blank); const [editing, setEditing] = useState(null); const [busy, setBusy] = useState(false);
  const reload = () => api("/categories", { auth: false }).then((data) => { window.dispatchEvent(new CustomEvent("catalog-refresh", { detail: { categories: data } })); }).catch(() => {});
  const submit = async (event) => {
    event.preventDefault(); setBusy(true);
    const body = { ...form, slug: form.slug || createSlug(form.name), parent_id: form.parent_id ? Number(form.parent_id) : null };
    try { await api(editing ? "/admin/categories/" + editing : "/admin/categories", { method: editing ? "PUT" : "POST", body }); setForm(blank); setEditing(null); await reload(); toast("دسته‌بندی ذخیره شد."); }
    catch (failure) { toast(failure.message, "error"); } finally { setBusy(false); }
  };
  const uploadHomeImage = async (file) => {
    if (!file) return;
    const data = new FormData(); data.append("file", file);
    try { const result = await api("/admin/uploads", { method: "POST", body: data }); setForm((item) => ({ ...item, home_image_url: result.url })); toast("تصویر دسته بارگذاری شد."); }
    catch (failure) { toast(failure.message, "error"); }
  };
  const remove = async (item) => { if (!confirm("دسته «" + item.name + "» حذف شود؟")) return; try { await api("/admin/categories/" + item.id, { method: "DELETE" }); await reload(); toast("دسته حذف شد."); } catch (failure) { toast(failure.message, "error"); } };
  return <div className="admin-content admin-two-column">
    <form className="admin-form-panel" onSubmit={submit}>
      <span className="eyebrow">ساختار فروشگاه</span><h2>{editing ? "ویرایش دسته" : "دسته جدید"}</h2>
      <label className="field"><span>نام دسته</span><input required value={form.name} onChange={(event) => setForm((item) => ({ ...item, name: event.target.value, slug: item.slug ? item.slug : createSlug(event.target.value) }))} /></label>
      <label className="field"><span>اسلاگ</span><input dir="ltr" value={form.slug} onChange={(event) => setForm((item) => ({ ...item, slug: createSlug(event.target.value) }))} /></label>
      <label className="field"><span>دسته والد</span><select value={form.parent_id || ""} onChange={(event) => setForm((item) => ({ ...item, parent_id: event.target.value, show_on_home: event.target.value ? false : item.show_on_home }))}><option value="">دسته اصلی</option>{categories.filter((item) => !item.parent_id && item.id !== editing).map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</select></label>
      <RichTextEditor label="توضیحات دسته" value={form.description || ""} onChange={(value) => setForm((item) => ({ ...item, description: value }))} placeholder="راهنمای انتخاب و معرفی این دسته برای مشتری…" minHeight={150} />
      <label className="field"><span>عنوان سئو</span><input value={form.seo_title || ""} onChange={(event) => setForm((item) => ({ ...item, seo_title: event.target.value }))} /></label>
      <label className="field"><span>توضیحات متا برای گوگل</span><textarea rows={2} maxLength={320} value={form.seo_description || ""} onChange={(event) => setForm((item) => ({ ...item, seo_description: event.target.value }))} /></label>
      <section className="admin-home-tile-fields"><h3>کارت دسته در صفحهٔ اول</h3>
        <label className="active-checkbox"><input type="checkbox" disabled={Boolean(form.parent_id)} checked={Boolean(form.show_on_home) && !form.parent_id} onChange={(event) => setForm((item) => ({ ...item, show_on_home: event.target.checked }))} /><span><b>نمایش در صفحهٔ اول</b><small>این گزینه برای دسته‌های اصلی فعال است؛ زیر‌دسته‌ها داخل صفحهٔ دسته‌بندی دیده می‌شوند.</small></span></label>
        <label className="field"><span>عنوان روی کارت</span><input value={form.home_title || ""} onChange={(event) => setForm((item) => ({ ...item, home_title: event.target.value }))} placeholder={form.name || "مثلاً زیورآلات روزمره"} /></label>
        <label className="field"><span>آدرس تصویر</span><input dir="ltr" value={form.home_image_url || ""} onChange={(event) => setForm((item) => ({ ...item, home_image_url: event.target.value }))} placeholder="https://… یا بارگذاری تصویر" /></label>
        <label className="upload-tile"><ImagePlus size={18} /><span>بارگذاری تصویر دسته</span><input type="file" accept="image/jpeg,image/png,image/webp" onChange={(event) => { uploadHomeImage(event.target.files?.[0]); event.target.value = ""; }} /></label>
        {form.home_image_url && <img className="home-tile-preview" src={form.home_image_url} alt="پیش‌نمایش تصویر دسته" />}
      </section>
      <button className="button button-red" disabled={busy}>{editing ? "ذخیره تغییرات" : "افزودن دسته"} <ArrowLeft size={15} /></button>
      {editing && <button type="button" className="text-link" onClick={() => { setEditing(null); setForm(blank); }}>انصراف از ویرایش</button>}
    </form>
    <div className="admin-list-panel"><div className="admin-panel-heading"><div><span className="eyebrow">چیدمان دسته‌ها</span><h2>دسته‌بندی‌ها</h2></div></div>
      <div className="simple-admin-list">{categories.map((item) => <div key={item.id} className="simple-admin-row"><span className="simple-row-icon">{item.parent_id ? "↳" : "▤"}</span><span><b>{item.name}</b><small>{item.slug}{item.parent_name ? " · زیرمجموعه " + item.parent_name : ""}{item.show_on_home ? " · کارت صفحهٔ اول" : ""}</small></span>
        <div className="row-actions"><button onClick={() => { setEditing(item.id); setForm({ ...blank, ...item, parent_id: item.parent_id ? String(item.parent_id) : "" }); }} aria-label="ویرایش"><PenIcon /></button><button onClick={() => remove(item)} aria-label="حذف"><Trash2 size={15} /></button></div>
      </div>)}{!categories.length && <p className="muted">دسته‌ای ثبت نشده است.</p>}</div>
    </div>
  </div>;
}

function AdminBrands({ brands, toast, formPage, editID }) {
  const blank = { name: "", slug: "", description: "", seo_title: "", seo_description: "" };
  const [form, setForm] = useState(blank); const [editing, setEditing] = useState(null);
  useEffect(() => {
    if (!formPage || !editID) { setEditing(null); setForm(blank); return; }
    const brand = brands.find((item) => String(item.id) === String(editID));
    if (brand) { setEditing(brand.id); setForm({ ...blank, ...brand }); }
  }, [formPage, editID, brands]);
  const editBrand = (brand) => { setEditing(brand.id); setForm({ ...blank, ...brand }); go(`/admin/brands/edit/${brand.id}`); };
  const submit = async (event) => {
    event.preventDefault();
    try { await api(editing ? "/admin/brands/" + editing : "/admin/brands", { method: editing ? "PUT" : "POST", body: { ...form, slug: form.slug || createSlug(form.name) } }); setForm(blank); setEditing(null); window.dispatchEvent(new Event("catalog-refresh")); toast("برند ذخیره شد."); go("/admin/brands"); }
    catch (failure) { toast(failure.message, "error"); }
  };
  const remove = async (brand) => { if (!confirm("برند «" + brand.name + "» حذف شود؟")) return; try { await api("/admin/brands/" + brand.id, { method: "DELETE" }); window.dispatchEvent(new Event("catalog-refresh")); toast("برند حذف شد."); } catch (failure) { toast(failure.message, "error"); } };
  return <div className={`admin-content admin-two-column admin-brands-page ${formPage ? 'admin-brand-form-page' : ''}`}>
    {!formPage && <div className="admin-list-panel"><div className="admin-panel-heading"><div><span className="eyebrow">ویترین برندها</span><h2>برندها <small>{fa(brands.length)}</small></h2></div><a className="button button-red button-small" href={routeHref('/admin/brands/new')}><Plus size={14} /> برند جدید</a></div><div className="brand-admin-grid">{brands.map((brand) => <article key={brand.id} className="brand-admin-card"><span className="brand-initial">{brand.name.slice(0, 1)}</span><b>{brand.name}</b><small>{brand.slug}</small><div className="row-actions"><button type="button" aria-label={`ویرایش ${brand.name}`} onClick={() => editBrand(brand)}><PenIcon /></button><button type="button" aria-label={`حذف ${brand.name}`} onClick={() => remove(brand)}><Trash2 size={15} /></button></div></article>)}</div>{!brands.length && <p className="muted">هنوز برندی ثبت نشده است.</p>}</div>}
    {formPage && <form className="admin-form-panel" onSubmit={submit}><div className="admin-panel-heading"><div><span className="eyebrow">برندهای انتخاب‌شده</span><h2>{editing ? "ویرایش برند" : "افزودن برند"}</h2></div><div className="admin-form-heading-actions"><a className="text-link" href={routeHref('/admin/brands')}>بازگشت به فهرست <ArrowRight size={14} /></a>{editing && <button className="icon-button" type="button" onClick={() => go('/admin/brands')} aria-label="لغو ویرایش"><X size={18} /></button>}</div></div>
      <label className="field"><span>نام برند</span><input required value={form.name} onChange={(event) => setForm((item) => ({ ...item, name: event.target.value, slug: item.slug ? item.slug : createSlug(event.target.value) }))} /></label>
      <label className="field"><span>اسلاگ</span><input dir="ltr" value={form.slug} onChange={(event) => setForm((item) => ({ ...item, slug: createSlug(event.target.value) }))} /></label>
      <RichTextEditor label="معرفی مفصل برند" value={form.description || ""} onChange={(value) => setForm((item) => ({ ...item, description: value }))} placeholder="تاریخچه، سبک و راهنمای انتخاب محصولات این برند…" minHeight={220} />
      <label className="field"><span>عنوان سئو</span><input value={form.seo_title || ""} onChange={(event) => setForm((item) => ({ ...item, seo_title: event.target.value }))} /></label>
      <label className="field"><span>توضیحات متا برای گوگل</span><textarea rows={2} maxLength={320} value={form.seo_description || ""} onChange={(event) => setForm((item) => ({ ...item, seo_description: event.target.value }))} /></label>
      <button className="button button-red">{editing ? "ذخیره تغییرات" : "افزودن برند"} <ArrowLeft size={15} /></button>
    </form>}
  </div>;
}

function AdminOrders({ toast }) {
  const [orders, setOrders] = useState([]); const [status, setStatus] = useState('');
  const load = () => api(`/admin/orders?page=1&limit=100${status ? `&status=${status}` : ''}`).then((result) => setOrders(result.data || [])).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, [status]);
  const changeStatus = async (order, next) => { try { await api(`/admin/orders/${order.id}/status`, { method: 'PATCH', body: { status: next } }); await load(); toast('وضعیت سفارش به‌روز شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className="admin-list-panel"><div className="admin-panel-heading"><div><span className="eyebrow">پردازش سفارش</span><h2>جریان سفارش‌ها</h2></div><select className="status-filter" value={status} onChange={(event) => setStatus(event.target.value)}><option value="">همه وضعیت‌ها</option>{Object.entries(ORDER_LABELS).map(([key, label]) => <option key={key} value={key}>{label}</option>)}</select></div><div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>شماره</th><th>مشتری</th><th>اقلام</th><th>مبلغ</th><th>وضعیت</th><th>عملیات</th></tr></thead><tbody>{orders.map((order) => <tr key={order.id}><td>#{fa(order.id)}<small>{shortDate(order.created_at)}</small></td><td>{order.customer_name}<small>{order.customer_phone}</small></td><td>{fa(order.items?.length || 0)} قلم</td><td>{money(order.total_amount)}</td><td><span className={`status-pill status-${order.status}`}>{ORDER_LABELS[order.status]}</span></td><td><div className="order-row-actions"><a className="button button-small button-outline" href={routeHref(`/admin/orders/${order.id}`)}><FileText size={14} />جزئیات</a>{order.status === 'paid' ? <button className="button button-small button-outline" onClick={() => changeStatus(order, 'processing')}>آماده‌سازی</button> : order.status === 'processing' ? <button className="button button-small button-red" onClick={() => changeStatus(order, 'shipped')}>ثبت ارسال</button> : order.status === 'awaiting_payment' ? <button className="text-link" onClick={() => changeStatus(order, 'cancelled')}>لغو سفارش</button> : null}</div></td></tr>)}</tbody></table></div>{!orders.length && <p className="muted table-empty">سفارشی برای این فیلتر ثبت نشده است.</p>}</div>;
}

function AdminOrderDetail({ id, toast }) {
  const [order, setOrder] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError('');
    api(`/admin/orders/${id}`)
      .then((result) => { if (active) setOrder(result); })
      .catch((failure) => {
        if (!active) return;
        setError(failure.message);
        toast(failure.message, 'error');
      })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [id, toast]);

  if (loading) return <LoadingState />;
  if (error || !order) return <div className="admin-order-detail"><a className="text-link" href={routeHref('/admin/orders')}><ArrowRight size={15} /> بازگشت به سفارش‌ها</a><ErrorPanel message={error || 'سفارش پیدا نشد.'} /></div>;

  const address = order.shipping_address || {};
  const receiverName = address.receiver_name || order.customer_name || '—';
  const receiverPhone = address.receiver_phone || order.customer_phone || '—';
  const addressParts = [address.province, address.city, address.address_line].filter(Boolean);

  return <section className="admin-order-detail">
    <div className="invoice-actions"><a className="button button-outline" href={routeHref('/admin/orders')}><ArrowRight size={16} /> بازگشت به سفارش‌ها</a><button className="button button-green" type="button" onClick={() => window.print()}><Printer size={16} /> چاپ فاکتور</button></div>
    <article className="order-invoice invoice-page" dir="rtl">
      <header className="invoice-header">
        <div><span className="eyebrow">گالری طلا و زیورآلات ملکی</span><h2>فاکتور سفارش</h2><p>زیبایی در جزئیات ماندگار است.</p></div>
        <div className="invoice-order-meta"><strong>شماره سفارش #{fa(order.id)}</strong><span>تاریخ ثبت: {dateTime(order.created_at)}</span><span className={`status-pill status-${order.status}`}>{ORDER_LABELS[order.status] || order.status}</span></div>
      </header>
      <div className="invoice-customer-grid">
        <section className="invoice-info-card"><h3><UserRound size={16} /> گیرنده</h3><b>{receiverName}</b><span dir="ltr">{receiverPhone}</span></section>
        <section className="invoice-info-card"><h3><MapPin size={16} /> نشانی ارسال</h3><p>{addressParts.length ? addressParts.join('، ') : 'نشانی برای این سفارش ثبت نشده است.'}</p>{address.postal_code && <span>کد پستی: <bdi dir="ltr">{address.postal_code}</bdi></span>}</section>
        <section className="invoice-info-card"><h3><Truck size={16} /> روش ارسال</h3><b>{order.shipping_method_name || '—'}</b>{order.tracking_code && <span>کد رهگیری: <bdi dir="ltr">{order.tracking_code}</bdi></span>}</section>
      </div>
      <div className="invoice-items-wrap"><table className="invoice-items"><thead><tr><th>ردیف</th><th>شرح کالا</th><th>تعداد</th><th>قیمت واحد</th><th>جمع</th></tr></thead>
        <tbody>{(order.items || []).map((item, index) => <tr key={`${item.product_id}-${item.variant_id || index}`}><td>{fa(index + 1)}</td><td><b>{item.product_name}</b>{item.variant_name && <small>{item.variant_label ? `${item.variant_label}: ` : ''}{item.variant_name}</small>}</td><td>{fa(item.quantity)}</td><td>{money(item.unit_price)}</td><td>{money(item.subtotal)}</td></tr>)}</tbody>
      </table></div>
      <div className="invoice-summary"><div><span>جمع کالاها</span><b>{money(order.items_amount)}</b></div>{Number(order.discount_amount) > 0 && <div><span>تخفیف{order.coupon_code ? ` (${order.coupon_code})` : ''}</span><b>− {money(order.discount_amount)}</b></div>}<div><span>هزینهٔ ارسال</span><b>{Number(order.shipping_cost) ? money(order.shipping_cost) : 'رایگان'}</b></div><div className="invoice-total"><strong>مبلغ نهایی</strong><strong>{money(order.total_amount)}</strong></div></div>
      <footer className="invoice-footer"><span>با سپاس از انتخاب شما</span><span>تاریخ چاپ: {dateTime(new Date().toISOString())}</span></footer>
    </article>
  </section>;
}

function AdminUsers({ toast }) {
  const [users, setUsers] = useState([]); const [search, setSearch] = useState('');
  const load = () => api(`/admin/users?page=1&limit=100${search ? `&q=${encodeURIComponent(search)}` : ''}`).then((result) => setUsers(result.data || [])).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, [search]);
  const access = async (user, field, value) => { const body = { role: field === 'role' ? value : user.role, is_active: field === 'active' ? value : user.is_active }; try { await api(`/admin/users/${user.id}/access`, { method: 'PATCH', body }); await load(); toast('دسترسی کاربر به‌روز شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className="admin-list-panel"><div className="admin-panel-heading"><div><span className="eyebrow">مشتری‌های MALEKI</span><h2>مدیریت کاربران</h2></div><label className="admin-search"><Search size={15} /><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="نام یا شماره موبایل" /></label></div><div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>کاربر</th><th>شماره تماس</th><th>عضویت</th><th>نقش</th><th>دسترسی</th></tr></thead><tbody>{users.map((user) => <tr key={user.id}><td><b>{user.first_name} {user.last_name}</b><small>#{fa(user.id)}</small></td><td dir="ltr">{user.phone}</td><td>{shortDate(user.created_at)}</td><td><select value={user.role} onChange={(event) => access(user, 'role', event.target.value)}><option value="user">کاربر</option><option value="admin">مدیر</option></select></td><td><button className={`visibility-switch ${user.is_active ? 'switch-on' : ''}`} onClick={() => access(user, 'active', !user.is_active)}><span />{user.is_active ? 'فعال' : 'مسدود'}</button></td></tr>)}</tbody></table></div></div>;
}

function AdminComments({ toast }) {
  const [comments, setComments] = useState([]); const [filter, setFilter] = useState('pending');
  const load = () => api(`/admin/comments?page=1&limit=100${filter ? `&status=${filter}` : ''}`).then((result) => setComments(result.data || [])).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, [filter]);
  const approve = async (item, value) => { try { await api(`/admin/comments/${item.id}/approval`, { method: 'PATCH', body: { is_approved: value } }); await load(); toast(value ? 'دیدگاه منتشر شد.' : 'نمایش دیدگاه متوقف شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  return <div className="admin-list-panel"><div className="admin-panel-heading"><div><span className="eyebrow">اعتماد مشتری</span><h2>دیدگاه‌های محصول</h2></div><select className="status-filter" value={filter} onChange={(event) => setFilter(event.target.value)}><option value="pending">در انتظار بررسی</option><option value="approved">منتشرشده</option><option value="">همه</option></select></div><div className="comment-admin-list">{comments.map((item) => <article className="comment-admin-card" key={item.id}><div><b>{item.author_name}</b><span>{item.product_name} · {shortDate(item.created_at)}</span></div><p>{item.body}</p><button className={`button button-small ${item.is_approved ? 'button-outline' : 'button-red'}`} onClick={() => approve(item, !item.is_approved)}>{item.is_approved ? 'توقف نمایش' : 'تایید و انتشار'}</button></article>)}{!comments.length && <p className="muted">دیدگاهی برای بررسی نیست.</p>}</div></div>;
}

function AdminBlog({ toast }) {
  const blank = { title: "", slug: "", summary: "", content: "", cover_image_url: "", seo_title: "", seo_description: "", is_published: false };
  const [posts, setPosts] = useState([]); const [editing, setEditing] = useState(null); const [form, setForm] = useState(blank);
  const load = () => api("/admin/blog/posts?page=1&limit=100").then((result) => setPosts(result.data || [])).catch((failure) => toast(failure.message, "error"));
  useEffect(() => { load(); }, []);
  const set = (key, value) => setForm((item) => ({ ...item, [key]: value }));
  const edit = async (item) => { try { const post = await api("/admin/blog/posts/" + item.id); setEditing(post.id); setForm({ ...post, is_published: Boolean(post.is_published) }); } catch (failure) { toast(failure.message, "error"); } };
  const submit = async (event) => {
    event.preventDefault();
    try { await api(editing ? "/admin/blog/posts/" + editing : "/admin/blog/posts", { method: editing ? "PUT" : "POST", body: { ...form, slug: form.slug || createSlug(form.title) } }); setEditing(null); setForm(blank); await load(); toast("مطلب ذخیره شد."); }
    catch (failure) { toast(failure.message, "error"); }
  };
  const remove = async (post) => { if (!confirm("مطلب «" + post.title + "» حذف شود؟")) return; try { await api("/admin/blog/posts/" + post.id, { method: "DELETE" }); await load(); toast("مطلب حذف شد."); } catch (failure) { toast(failure.message, "error"); } };
  return <div className="admin-content admin-blog-layout"><form className="admin-form-panel" onSubmit={submit}><span className="eyebrow">محتوای فروشگاه</span><h2>{editing ? "ویرایش مطلب" : "مطلب تازه"}</h2>
    <label className="field"><span>عنوان</span><input required value={form.title} onChange={(event) => setForm((item) => ({ ...item, title: event.target.value, slug: item.slug ? item.slug : createSlug(event.target.value) }))} /></label>
    <label className="field"><span>اسلاگ</span><input dir="ltr" value={form.slug} onChange={(event) => set("slug", createSlug(event.target.value))} /></label>
    <label className="field"><span>خلاصه</span><textarea rows={2} value={form.summary} onChange={(event) => set("summary", event.target.value)} /></label>
    <RichTextEditor label="متن وبلاگ" value={form.content || ""} onChange={(value) => set("content", value)} placeholder="مطلب را با تیترها، فهرست و لینک بنویس…" minHeight={300} />
    <label className="field"><span>تصویر شاخص (نشانی)</span><input dir="ltr" value={form.cover_image_url || ""} onChange={(event) => set("cover_image_url", event.target.value)} /></label>
    <label className="field"><span>عنوان SEO</span><input value={form.seo_title || ""} onChange={(event) => set("seo_title", event.target.value)} /></label>
    <label className="field"><span>توضیحات متا</span><textarea rows={2} maxLength={320} value={form.seo_description} onChange={(event) => set("seo_description", event.target.value)} /></label>
    <label className="active-checkbox"><input type="checkbox" checked={Boolean(form.is_published)} onChange={(event) => set("is_published", event.target.checked)} /><span><b>انتشار عمومی</b><small>صفحهٔ مطلب با metadata سئو در دسترس خواهد بود.</small></span></label>
    <button className="button button-red">ذخیره مطلب <ArrowLeft size={15} /></button>{editing && <button type="button" className="text-link" onClick={() => { setEditing(null); setForm(blank); }}>انصراف از ویرایش</button>}
  </form>
  <div className="admin-list-panel"><span className="eyebrow">مجله MALEKI</span><h2>مطالب و راهنماها</h2><div className="simple-admin-list">{posts.map((post) => <div className="simple-admin-row" key={post.id}><span className="simple-row-icon"><PenIcon /></span><span><b>{post.title}</b><small>{post.is_published ? "منتشرشده" : "پیش‌نویس"} · {post.slug}</small></span><div className="row-actions"><button onClick={() => edit(post)}><PenIcon /></button><button onClick={() => remove(post)}><Trash2 size={15} /></button>{post.is_published && <a href={"/blog/" + post.slug} target="_blank" rel="noreferrer"><ArrowUpLeft size={14} /></a>}</div></div>)}</div></div></div>;
}

function AdminSiteContent({ toast }) {
  const blank = { contact_phone: "", contact_email: "", contact_address: "", contact_hours: "", instagram_url: "", about_content: "", terms_content: "" };
  const [form, setForm] = useState(blank); const [loading, setLoading] = useState(true); const [busy, setBusy] = useState(false);
  useEffect(() => { let active = true; api("/site-content", { auth: false }).then((content) => { if (active) setForm({ ...blank, ...content }); }).catch((failure) => toast(failure.message, "error")).finally(() => { if (active) setLoading(false); }); return () => { active = false; }; }, []);
  const set = (key, value) => setForm((item) => ({ ...item, [key]: value }));
  const save = async (event) => { event.preventDefault(); setBusy(true); try { await api("/admin/site-content", { method: "PUT", body: form }); toast("اطلاعات صفحات ذخیره شد."); } catch (failure) { toast(failure.message, "error"); } finally { setBusy(false); } };
  if (loading) return <LoadingState />;
  return <form className="admin-content admin-form-panel admin-site-content-form" onSubmit={save}><span className="eyebrow">اطلاعات عمومی فروشگاه</span><h2>تماس، درباره ما و قوانین</h2>
    <div className="form-grid">
      <label className="field"><span>شمارهٔ موبایل / تماس</span><input dir="ltr" value={form.contact_phone} onChange={(event) => set("contact_phone", event.target.value)} placeholder="09…" /></label>
      <label className="field"><span>ایمیل</span><input dir="ltr" type="email" value={form.contact_email} onChange={(event) => set("contact_email", event.target.value)} /></label>
      <label className="field field-wide"><span>نشانی</span><input value={form.contact_address} onChange={(event) => set("contact_address", event.target.value)} /></label>
      <label className="field"><span>ساعت پاسخ‌گویی</span><input value={form.contact_hours} onChange={(event) => set("contact_hours", event.target.value)} placeholder="شنبه تا پنجشنبه، ۹ تا ۱۸" /></label>
      <label className="field"><span>اینستاگرام</span><input dir="ltr" type="url" value={form.instagram_url} onChange={(event) => set("instagram_url", event.target.value)} placeholder="https://instagram.com/…" /></label>
    </div>
    <RichTextEditor label="متن درباره ما" value={form.about_content} onChange={(value) => set("about_content", value)} placeholder="داستان فروشگاه، رویکرد و خدمات…" minHeight={240} />
    <RichTextEditor label="قوانین و مقررات" value={form.terms_content} onChange={(value) => set("terms_content", value)} placeholder="قوانین خرید، ارسال، مرجوعی و حریم خصوصی…" minHeight={300} />
    <button className="button button-red" disabled={busy}>{busy ? "در حال ذخیره…" : "ذخیره محتوای سایت"} <ArrowLeft size={15} /></button>
  </form>;
}

function analyticsPageName(path) {
  if (path === "/") return "صفحهٔ اصلی";
  if (path === "/shop") return "فروشگاه";
  if (path.startsWith("/products/")) return "صفحهٔ محصول";
  if (path.startsWith("/categories/")) return "دسته‌بندی";
  if (path.startsWith("/brands/")) return "برند";
  if (path.startsWith("/account")) return "حساب کاربری";
  if (path.startsWith("/checkout")) return "پرداخت";
  if (path.startsWith("/contact")) return "تماس با ما";
  if (path.startsWith("/blog")) return "وبلاگ";
  if (path.startsWith("/admin")) return "مدیریت";
  return path;
}

function AdminAnalytics({ toast }) {
  const [data, setData] = useState(null); const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    const load = () => api("/admin/analytics").then((value) => { if (active) { setData(value); setError(""); } }).catch((failure) => { if (active) setError(failure.message); });
    load(); const timer = window.setInterval(load, 10000);
    return () => { active = false; window.clearInterval(timer); };
  }, []);
  if (error && !data) return <div className="admin-content"><ErrorPanel message={error} /></div>;
  if (!data) return <LoadingState />;
  const metrics = [
    ["آنلاین همین لحظه", fa(data.active_online), "heartbeat در ۹۰ ثانیهٔ اخیر", "◉"],
    ["بازدیدکنندهٔ امروز", fa(data.visitors_today), "شناسه‌های ناشناس روز جاری", "♙"],
    ["بازدیدکنندهٔ کل", fa(data.unique_visitors), "شناسه‌های ناشناس در کل دوره", "♧"],
    ["بازدید صفحه امروز", fa(data.page_views_today), "نمایش‌های ثبت‌شده امروز", "▤"],
  ];
  return <div className="admin-content admin-analytics"><div className="admin-stats">{metrics.map(([label, value, hint, icon]) => <article className="admin-stat-card" key={label}><div className="admin-stat-heading"><small>{label}</small><span className="admin-stat-icon" aria-hidden="true">{icon}</span></div><strong>{value}</strong><span>{hint}</span></article>)}</div>
    {error && <ErrorPanel message={error} />}
    <section className="admin-panel"><div className="admin-panel-heading"><div><span className="eyebrow">تازه‌سازی خودکار هر ۱۰ ثانیه</span><h2>بازدیدکنندگان فعال و صفحهٔ جاری</h2></div><span className="status-pill status-paid">{fa(data.active_online)} آنلاین</span></div>
      {data.active_visitors?.length ? <div className="admin-table-wrap"><table className="admin-table"><thead><tr><th>شناسهٔ ناشناس</th><th>صفحه</th><th>مسیر</th><th>آخرین فعالیت</th></tr></thead><tbody>{data.active_visitors.map((visitor) => <tr key={visitor.visitor_id}><td dir="ltr">{visitor.visitor_id.slice(0, 8)}…</td><td>{analyticsPageName(visitor.path)}</td><td dir="ltr">{visitor.path}</td><td>{dateTime(visitor.last_seen)}</td></tr>)}</tbody></table></div> : <p className="muted">بازدیدکنندهٔ فعالی ثبت نشده؛ این صفحه با heartbeat مرورگرها به‌روز می‌شود.</p>}
    </section><p className="muted analytics-privacy-note">برای شمارش بازدید، یک شناسهٔ تصادفی در مرورگر ساخته می‌شود. نام، شمارهٔ موبایل و نشانی IP در این بخش ثبت نمی‌شود.</p>
  </div>;
}

function AdminTickets({ toast }) {
  const [tickets, setTickets] = useState([]); const [status, setStatus] = useState('');
  useEffect(() => { api(`/admin/tickets?page=1&limit=100${status ? `&status=${status}` : ''}`).then((result) => setTickets(result.data || [])).catch((failure) => toast(failure.message, 'error')); }, [status]);
  return <div className="admin-list-panel"><div className="admin-panel-heading"><div><span className="eyebrow">پشتیبانی MALEKI</span><h2>درخواست‌های کاربران</h2></div><select className="status-filter" value={status} onChange={(event) => setStatus(event.target.value)}><option value="">همهٔ تیکت‌ها</option><option value="open">باز</option><option value="answered">پاسخ‌داده‌شده</option><option value="closed">بسته</option></select></div><div className="admin-ticket-list">{tickets.map((ticket) => <a className="admin-ticket-row" href={routeHref(`/admin/tickets/${ticket.id}`)} key={ticket.id}><span className="ticket-row-icon"><MessageIcon /></span><span><b>{ticket.subject}</b><small>{ticket.user_name} · {ticket.user_phone}</small><small>{ticket.last_message}</small></span><span className={`status-pill status-${ticket.status}`}>{ticket.status === 'answered' ? 'پاسخ داده شد' : ticket.status === 'closed' ? 'بسته' : 'باز'}</span><ArrowLeft size={15} /></a>)}</div>{!tickets.length && <p className="muted">تیکتی برای این وضعیت نیست.</p>}</div>;
}

function AdminTicketDetail({ id, toast }) {
  const [ticket, setTicket] = useState(null); const [body, setBody] = useState('');
  const load = () => api(`/admin/tickets/${id}`).then(setTicket).catch((failure) => toast(failure.message, 'error'));
  useEffect(() => { load(); }, [id]);
  const reply = async (event) => { event.preventDefault(); try { await api(`/admin/tickets/${id}/messages`, { method: 'POST', body: { body } }); setBody(''); await load(); toast('پاسخ ارسال شد.'); } catch (failure) { toast(failure.message, 'error'); } };
  const status = async (value) => { try { await api(`/admin/tickets/${id}/status`, { method: 'PATCH', body: { status: value } }); await load(); toast('وضعیت تیکت تغییر کرد.'); } catch (failure) { toast(failure.message, 'error'); } };
  if (!ticket) return <LoadingState />;
  return <div className="admin-ticket-detail"><a className="back-link" href={routeHref('/admin/tickets')}><ArrowRight size={15} /> بازگشت به تیکت‌ها</a><div className="admin-ticket-head"><div><span className="eyebrow">{ticket.user_name} · {ticket.user_phone}</span><h2>{ticket.subject}</h2></div><div className="ticket-status-actions"><button className="button button-small button-outline" onClick={() => status(ticket.status === 'closed' ? 'open' : 'closed')}>{ticket.status === 'closed' ? 'بازگشایی' : 'بستن تیکت'}</button></div></div>
    <div className="conversation-messages">{(ticket.messages || []).map((message) => <article className={`message-bubble ${message.sender_role === 'admin' ? 'message-support' : 'message-customer'}`} key={message.id}><div><b>{message.sender_role === 'admin' ? 'پشتیبانی MALEKI' : message.sender_name}</b><time>{dateTime(message.created_at)}</time></div><p>{message.body}</p></article>)}</div>
    {ticket.status !== 'closed' && <form className="reply-form" onSubmit={reply}><textarea value={body} onChange={(event) => setBody(event.target.value)} required rows={4} placeholder="پاسخ مدیر…" /><button className="button button-red">ارسال پاسخ <Send size={15} /></button></form>}</div>;
}
