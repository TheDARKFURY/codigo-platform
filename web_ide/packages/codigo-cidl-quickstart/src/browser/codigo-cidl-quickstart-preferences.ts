import {PreferenceSchema} from '@theia/core/lib/common/preferences/preference-schema';

export namespace CodigoPreferences {
    export const alwaysShowQuickStartPage = 'codigo.alwaysShowQuickStart';
}

export const codigoPreferenceSchema: PreferenceSchema = {
    'type': 'object',
    'properties': {
        'codigo.alwaysShowQuickStart': {
            type: 'boolean',
            description: 'Show QuickStart Page after every start of Código Studio.',
            default: true
        }
    }
};
