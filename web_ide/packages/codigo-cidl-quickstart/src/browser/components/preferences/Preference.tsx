import * as React from "react";
import {CodigoPreferences} from "../../codigo-cidl-quickstart-preferences";
import {PreferenceService} from "@theia/core/lib/browser";

export interface Props {
    preferenceService: PreferenceService;
}

const Preference = (props: Props) => {
    const [alwaysShowQuickStartPage, setAlwaysShowWelcomePage] = React.useState<boolean>(props.preferenceService.get(CodigoPreferences.alwaysShowQuickStartPage, true));

    React.useEffect(() => {
        const preflistener = props.preferenceService.onPreferenceChanged(change => {
            if (change.preferenceName === CodigoPreferences.alwaysShowQuickStartPage) {
                const prefValue: boolean = change.newValue;
                setAlwaysShowWelcomePage(prefValue);
            }
        });

        return () => preflistener.dispose();
    }, [props.preferenceService]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newChecked = e.target.checked;
        props.preferenceService.updateValue(CodigoPreferences.alwaysShowQuickStartPage, newChecked);
    };

    return (
        <div className="flex items-center mb-4">
            <input id="checkbox-1" type="checkbox" value="" onChange={handleChange}
                   checked={alwaysShowQuickStartPage}
                   className="w-4 h-4 rounded focus:ring-blue-600 ring-offset-gray-800 focus:ring-offset-gray-800 focus:ring-2 bg-gray-700 border-gray-600"/>
            <label htmlFor="checkbox-1" className="ml-2 text-sm font-medium text-gray-300">
                Show QuickStart Page after every start of Código Studio
            </label>
        </div>
    );
}

export {Preference};