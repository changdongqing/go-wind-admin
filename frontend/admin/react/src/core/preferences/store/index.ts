import {create} from 'zustand';
import {persist} from 'zustand/middleware';

import {defaultPreferences} from '../config/default';
import type {DeepPartial, Preferences} from '../types';
import {mergeDeep} from "../utils/merge";


export interface PreferencesState {
    preferences: Preferences;
    setPreferences: (overrides: DeepPartial<Preferences>) => void;
    resetPreferences: () => void;
    getPreference: <K extends keyof Preferences>(key: K) => Preferences[K];
}

export const usePreferencesStore = create<PreferencesState>()(
    persist(
        (set, get) => ({
            preferences: defaultPreferences,

            setPreferences: (overrides) => {
                set((state) => ({
                    preferences: mergeDeep(state.preferences, overrides),
                }));
            },

            resetPreferences: () => {
                set({preferences: defaultPreferences});
            },

            getPreference: (key) => {
                return get().preferences[key];
            },
        }),
        {
            name: 'app-preferences',
            partialize: (state) => ({preferences: state.preferences}),
            // 部署级运行时配置（public/app-config.js）优先级须高于浏览器缓存：
            // 否则切换客户/改名后，旧 appName 会被 localStorage 持久化的值覆盖，导致改名不生效
            merge: (persistedState, currentState) => {
                const persistedPrefs = (persistedState as { preferences?: DeepPartial<Preferences> })?.preferences ?? {};
                // 先并入用户持久化的偏好（主题/布局等），再用部署级运行时标识强制覆盖
                let merged = mergeDeep(currentState.preferences, persistedPrefs);
                const runtime = window.__APP_CONFIG__;
                if (runtime) {
                    const overrides: DeepPartial<Preferences> = {};
                    if (runtime.appName) {
                        overrides.app = { name: runtime.appName };
                    }
                    if (
                        runtime.copyrightCompanyName ||
                        runtime.copyrightYear ||
                        runtime.copyrightSiteLink ||
                        runtime.copyrightIcp
                    ) {
                        overrides.copyright = {
                            ...(runtime.copyrightCompanyName && { companyName: runtime.copyrightCompanyName }),
                            ...(runtime.copyrightYear && { date: runtime.copyrightYear }),
                            ...(runtime.copyrightSiteLink && { companySiteLink: runtime.copyrightSiteLink }),
                            ...(runtime.copyrightIcp && { icp: runtime.copyrightIcp }),
                        };
                    }
                    if (Object.keys(overrides).length > 0) {
                        merged = mergeDeep(merged, overrides);
                    }
                }
                return { ...currentState, preferences: merged };
            },
        }
    )
);
