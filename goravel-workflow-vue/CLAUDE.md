# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Vue 3 + TypeScript** frontend scaffolding for the `goravel-workflow` Go backend. It serves as the admin panel and workflow management UI, built with Vite, Ant Design Vue, VXE-Table, and jsPlumb for flowchart design. The project also includes device tracking, Excel import/export, command control, and socket features that belong to the parent application (not the workflow package itself).

## Tech Stack

- **Framework**: Vue 3.4 with Composition API (`<script setup>`)
- **Language**: TypeScript 5.3
- **Build Tool**: Vite 5
- **UI Library**: Ant Design Vue 4.1
- **Data Grid**: VXE-Table 4.5
- **State Management**: Pinia 2.1
- **Routing**: Vue Router 4.2 (hash mode via `createWebHashHistory`)
- **HTTP Client**: Axios 1.6 (custom wrapper class)
- **Styling**: Tailwind CSS 3.4 + SCSS + custom theme variables
- **Flowchart**: jsPlumb 2.15 (visual workflow designer)
- **Utilities**: dayjs, lodash-es, xe-utils, echarts, vue-draggable-plus

## Directory Structure

```
goravel-workflow-vue/
├── src/
│   ├── api/                    # Direct API calls (auth, socket, home)
│   │   ├── auth/useAuth.ts     # Login/register/forgot-password APIs
│   │   ├── home/useHome.ts     # Dashboard/home data
│   │   └── socket/useSocket.ts # WebSocket notifications
│   ├── assets/                 # Static assets (CSS, SCSS, images, icons, logos)
│   ├── components/             # Reusable UI components
│   │   ├── form/index.vue      # Reusable dynamic form component
│   │   ├── upload/index.vue    # File upload with /api/upload
│   │   ├── empsearch/index.vue # Employee search dropdown
│   │   ├── userlist/index.vue  # User selection list
│   │   ├── hulu-menu/index.vue # Right-click context menu for flow design
│   │   ├── attrform/index.vue  # Dynamic attribute form (process config)
│   │   └── board/              # Dashboard visualization components
│   ├── composables/            # Composable functions (API layer)
│   │   ├── bus/                # Business API callers
│   │   │   ├── useDept.ts      # Department CRUD + bind manager/director
│   │   │   ├── useEmp.ts       # Employee CRUD + search + options + bind user
│   │   │   ├── useEntry.ts     # Workflow entry operations (store/show/resend)
│   │   │   ├── useFlow.ts      # Flow CRUD + design + publish + templateform
│   │   │   ├── useFlowlink.ts  # Flowlink update (save node positions)
│   │   │   ├── useProc.ts      # Approval actions (pass/unpass/indexProcs)
│   │   │   ├── useProcess.ts   # Process step config (attribute/con/list)
│   │   │   ├── useTemplate.ts  # Template CRUD
│   │   │   ├── useTemplateForm.ts # Template form field CRUD
│   │   │   ├── useUser.ts      # User CRUD (parent app, not workflow package)
│   │   │   ├── usePlugin.ts    # Plugin install/uninstall
│   │   │   ├── usePluginConfig.ts # Plugin configuration APIs
│   │   │   ├── useCommand.ts   # Device command control (parent app)
│   │   │   ├── useExcel.ts     # Excel import/export (parent app)
│   │   │   └── useTrack.ts     # Device history track (parent app)
│   │   └── common/
│   │       ├── useStorage.ts   # localStorage wrapper
│   │       └── useUtil.ts      # Utility functions
│   ├── enum/                   # TypeScript enums
│   │   ├── ApiEnum.ts          # API endpoint constants
│   │   ├── RouteEnum.ts        # Route name constants
│   │   ├── RouteName.ts        # Route name aliases
│   │   ├── HttpCodeEnum.ts     # HTTP status codes
│   │   ├── HttpStatus.ts       # Business status codes
│   │   ├── CacheKey.ts         # localStorage key constants
│   │   └── CacheEnum.ts        # Cache key enums
│   ├── layouts/                # Page layout templates
│   │   ├── admin/              # Admin panel layout (sidebar, navbar, submenu)
│   │   ├── auth/               # Auth page layout (login/register)
│   │   └── error/              # Error page layout
│   ├── middleware/
│   │   └── appMiddleware.ts    # App-level middleware
│   ├── plugins/                # Plugin registration (auto-installed in main.ts)
│   │   ├── axios/Axios.ts      # Custom Axios wrapper with interceptors
│   │   ├── axios/index.ts      # Axios instance setup (baseURL: /api)
│   │   ├── router/index.ts     # Router setup + prefix handling
│   │   ├── router/guard.ts     # Navigation guards (auth, permissions)
│   │   ├── antdvue/index.ts    # Ant Design Vue full registration
│   │   ├── pinia/index.ts      # Pinia setup
│   │   ├── vxe-table/index.ts  # VXE-Table setup + custom theme
│   │   ├── tailwindcss/index.ts # Tailwind CSS setup
│   │   ├── dayjs/index.ts      # Day.js setup
│   │   ├── antIcons/index.ts   # Ant Design icons auto-import
│   │   ├── vue3Lottie/index.ts # Lottie animation plugin
│   │   ├── cropper/index.ts    # Image cropper plugin
│   │   ├── terminal/index.ts   # Web terminal plugin
│   │   ├── menucontext/index.ts # Context menu plugin
│   │   ├── gocaptchavue/index.ts # CAPTCHA plugin
│   │   └── provider/loading.ts # Loading state provider
│   ├── routes/                 # Route definitions
│   │   ├── index.ts            # Aggregates all route modules
│   │   ├── admin.ts            # Admin routes (dashboard, base management, task, execute)
│   │   ├── auth.ts             # Auth routes (login, register, forgot password)
│   │   ├── error.ts            # Error routes (404, 403, 500)
│   │   ├── default.ts          # Default redirect routes
│   │   ├── DynamicRoutes.ts    # Permission-based dynamic route loading
│   │   └── StaticRoutes.ts     # Static route constants
│   ├── store/                  # Pinia stores
│   │   ├── useUserStore.ts     # User authentication state
│   │   ├── useMenuStore.ts     # Menu/navigation state
│   │   ├── useLoadingStore.ts  # Global loading state
│   │   ├── useErrorStore.ts    # Error state
│   │   ├── useTipsStore.ts     # Toast/notification state
│   │   ├── useRulesStore.ts    # Form validation rules state
│   │   └: permissons.ts        # Permission/route guard logic
│   └── views/                  # Page components
│       ├── auth/               # Login, register, forgot password
│       ├── errors/             # 404 error page
│       ├── admin/
│       │   ├── Translation.vue # Nested route placeholder (translation)
│       │   ├── base/           # Core workflow management
│       │   │   ├── dashboard/  # Admin dashboard with Lottie animations
│       │   │   ├── dept/       # Department management (hierarchy, map, track)
│       │   │   ├── emp/        # Employee management
│       │   │   ├── user/       # User management
│       │   │   ├── flow/       # Workflow design (index, create, edit, design, initiation, proc)
│       │   │   ├── template/   # Form template management
│       │   │   ├── template_form/ # Template field management
│       │   │   ├── plugin/     # Plugin management
│       │   │   └── pdf/        # PDF generation
│       │   ├── task/           # Task management (dispatch, receive, import, subdivide)
│       │   └── execute/        # Execution management (batch, special)
├── core/                       # Core build configuration
├── public/                     # Public static assets
├── types/                      # TypeScript type declarations
├── vite.config.ts              # Vite build configuration
├── tailwind.config.js          # Tailwind CSS configuration
├── postcss.config.js           # PostCSS configuration
├── tsconfig.json               # Root TypeScript config (references node + app)
├── tsconfig.app.json           # App TypeScript config
├── tsconfig.node.json          # Node TypeScript config
├── env.d.ts                    # Environment type declarations
├── components.d.ts             # Auto-generated component declarations
├── package.json                # Dependencies and scripts
├── .env                        # Environment variables (VITE_API_URL)
└── pnpm-lock.yaml              # pnpm lockfile
```

## Architecture

### Request Flow

1. **Axios Instance** (`plugins/axios/`): Custom `Axios` class wraps axios with:
   - Base URL: `/api`
   - Request interceptor: Adds JWT `Authorization: Bearer <token>` header
   - Response interceptor: Shows success messages, handles error codes (401→login, 422→form rules, 403→error toast, 404→home, 500→detailed error)
   - Upload method: `http.Upload(url, data)` for multipart/form-data

2. **Composables** (`composables/bus/`): Each composable encapsulates all API calls for one domain:
   - Returns reactive data (VXE-Table gridOptions) + API methods
   - Uses relative URLs (e.g., `flow`, `emp`, `proc/index`) resolved against `/api` base
   - Pattern: `const { loadItems, storeItem } = useXxx()`

3. **Routes** (`routes/`): Hash-based routing with `/admin` prefix applied at setup time via `config.router.prefix`. Admin routes are nested under `layouts/admin/index.vue`, auth routes under `layouts/auth/index.vue`.

4. **Stores** (`store/`): Pinia stores for shared state — user auth, loading spinner, menus, form validation rules, notifications.

### Key Patterns

- **All API calls use relative paths** — the Axios base URL `/api` is prepended automatically
- **VXE-Table grids** are pre-configured with proxyConfig (remote sorting/filtering/pagination) in each composable
- **Dynamic forms** via the `<Form>` component (`components/form/index.vue`) accepts `fields` prop and emits `submit`
- **Flow designer** (`views/admin/base/flow/design.vue`) uses jsPlumb for visual workflow editing with drag-and-drop nodes
- **Plugins** are auto-registered in `plugins/index.ts` — new plugins must be added there

## Common Commands

```bash
# Development server (hot-reload on localhost:5173)
pnpm dev

# Type checking (Vue + TypeScript)
pnpm type-check

# Production build (outputs to dist/)
pnpm build

# Preview production build
pnpm preview

# Build only (no type checking)
pnpm build-only
```

## API Endpoint Reference

The frontend calls these endpoints (all prefixed with `/api`):

### Workflow Package Endpoints (from goravel-workflow backend)

| Category | Method | Path | Used By | Notes |
|----------|--------|------|---------|-------|
| Auth | POST | `/api/auth/login` | login.vue | Admin login |
| Auth | POST | `/api/h5/login` | — | H5 login (backend has it) |
| Captcha | GET | `/api/captcha/get` | login.vue | |
| Captcha | POST | `/api/captcha/validate` | — | |
| Upload | POST | `/api/upload` | upload/index.vue | Multipart |
| Home | GET | `/api/home` | useHome.ts | |
| Dept | GET | `/api/dept` | useDept.ts | Resource CRUD |
| Dept | POST | `/api/dept/bindmanager` | useDept.ts | |
| Dept | GET | `/api/dept/list` | useDept.ts | Flat list |
| Dept | POST | `/api/dept/binddirector` | useDept.ts | |
| Emp | GET | `/api/emp` | useEmp.ts | Resource CRUD |
| Emp | POST | `/api/emp` | useEmp.ts | |
| Emp | PUT | `/api/emp/{id}` | useEmp.ts | |
| Emp | DELETE | `/api/emp/{id}` | useEmp.ts | |
| Emp | POST | `/api/emp/search` | useEmp.ts | Search by name |
| Emp | GET | `/api/emp/options` | useEmp.ts | Options dropdown |
| Emp | POST | `/api/emp/bind` | useEmp.ts | Bind user |
| Flow | GET | `/api/flow` | useFlow.ts | Resource CRUD + grid |
| Flow | POST | `/api/flow` | useFlow.ts | |
| Flow | PUT | `/api/flow/{id}` | useFlow.ts | |
| Flow | GET | `/api/flow/list` | useFlow.ts | Flat list |
| Flow | GET | `/api/flow/create` | useFlow.ts | Create form |
| Flow | GET | `/api/flow/flowchart/{id}` | useFlow.ts | Designer data |
| Flow | POST | `/api/flow/publish` | useFlow.ts | |
| Flow | GET | `/api/flow/{id}/entry` | useEntry.ts | Entry form config |
| Entry | POST | `/api/entry` | useEntry.ts | Submit entry |
| Entry | GET | `/api/entry/{id}` | useEntry.ts | Show entry |
| Entry | GET | `/api/entry/{id}/entrydata` | useEntry.ts | Form data |
| Entry | POST | `/api/entry/resend` | useEntry.ts | Resend/retry |
| Entry | POST | `/api/entry/revoke` | — | Withdraw (backend has it) |
| Flowlink | POST | `/api/flowlink` | useFlowlink.ts | Save node positions |
| Template | GET | `/api/template` | useTemplate.ts | Resource CRUD |
| Template | POST | `/api/template` | useTemplate.ts | |
| Template | PUT | `/api/template/{id}` | useTemplate.ts | |
| Template | DELETE | `/api/template/{id}` | useTemplate.ts | |
| TemplateForm | GET | `/api/template/{id}/templateform` | useTemplateForm.ts | |
| TemplateForm | POST | `/api/templateform` | useTemplateForm.ts | |
| TemplateForm | PUT | `/api/templateform/{id}` | useTemplateForm.ts | |
| TemplateForm | DELETE | `/api/templateform/{id}` | useTemplateForm.ts | |
| TemplateForm | POST | `/api/flow/templateform` | useFlow.ts | Flow's templateforms |
| Process | GET | `/api/process/attribute` | useProcess.ts | Step attributes |
| Process | POST | `/api/process/con` | useProcess.ts | Conditions |
| Process | POST | `/api/process/list` | useProcess.ts | Flow's processes |
| Process | Resource CRUD | `/api/process` | — | Via router.Resource |
| Proc | GET | `/api/proc/{entry_id}` | useProc.ts | Frontend calls `proc/index?entry_id=...` |
| Proc | POST | `/api/pass` | useProc.ts | Approve |
| Proc | POST | `/api/unpass` | useProc.ts | Reject |
| Proc | POST | `/api/revoke` | — | Proc revoke (unused) |
| Proc | POST | `/api/addsign` | — | Add signer (unused) |
| Proc | POST | `/api/transfer` | — | Transfer (unused) |
| Proc | POST | `/api/comment` | — | Comment (unused) |
| Proc | GET | `/api/comments/{entry_id}` | — | Get comments (unused) |
| CC | GET | `/api/cc/list` | — | CC list (unused) |
| CC | GET | `/api/cc/entry/{entry_id}` | — | Entry CC records (unused) |

### Parent App Endpoints (NOT part of workflow package)

These endpoints are called by the frontend but are defined in the host application, not in the goravel-workflow package:

| Path | Used By | Notes |
|------|---------|-------|
| `/api/user` (CRUD) | useUser.ts | User management |
| `/api/device` (CRUD) | useTrack.ts | Device management |
| `/api/device/{id}/historytrack` | useTrack.ts | Device history |
| `/api/excel/preview_write` | useExcel.ts | Excel import |
| `/api/excel/preview_read` | useExcel.ts | Excel export |
| `/api/excel/reset/{id}` | useExcel.ts | |
| `/api/excel/fill` | useExcel.ts | |
| `/api/command/read/*` | useCommand.ts | Command control |
| `/api/plugin/list` | usePlugin.ts | Plugin listing |
| `/api/plugin/store_plugin_config` | usePluginConfig.ts | |
| `/api/plugin/get_plugin_config` | usePluginConfig.ts | |
| `/api/plugin/getall_plugin_config` | usePluginConfig.ts | |
| `/api/send_to_system` | useSocket.ts | WebSocket push |

### Discrepancies to Note

1. **Missing backend routes**: `plugin/*`, `user/*`, `device/*`, `excel/*`, `command/*`, `send_to_system` endpoints are NOT in `goravel-workflow/routes/api.go` — they are expected to be defined in the host application's routes.
2. **Unused backend endpoints**: Several backend endpoints exist but are not called by the current frontend (e.g., `process/con` for condition editing, `entry/{id}` for entry detail view).
3. **Dead code**: Previously many composables declared `serveApiUrl = import.meta.env.VITE_API_URL` but never use it — cleaned up in all 12 composables.

## Configuration

- **Router prefix**: Set in `src/config.ts` → `router.prefix` — prepended to all admin routes
- **API base URL**: Hardcoded as `/api` in `plugins/axios/index.ts`
- **External API URL**: `.env` → `VITE_API_URL` — referenced in some files but largely unused (dead code)
- **Route mode**: Hash-based (`createWebHashHistory`) — no server URL rewriting needed
- **Token storage**: Key from `CacheEnum.TOKEN_NAME` in localStorage

## Important Notes

1. **This is a scaffold/template project** — it includes both workflow-specific pages AND parent-app pages (device, excel, command). Only the `views/admin/base/` workflow pages correspond to the goravel-workflow backend package.

2. **VXE-Table gridOptions are shared patterns** — every composable defines a nearly identical `gridOptions` object with toolbar, proxy config, and columns. These are copy-paste templates that should be customized per page.

3. **The flow designer** (`design.vue` + `flow.ts`) is the most complex UI component — it renders a jsPlumb canvas for visual workflow editing with draggable nodes, connections, and a settings modal (`attrform`).

4. **No test framework configured** — there are no test files or test scripts in `package.json`.

5. **Component auto-imports** — `unplugin-vue-components` and `unplugin-auto-import` handle automatic imports. Check `components.d.ts` for the generated declarations.
