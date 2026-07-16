// Stub out @vue/devtools-kit which tries to access localStorage at import time
vi.mock('@vue/devtools-kit', () => ({ default: {}, DevToolsTimelineLayersState: undefined }))
