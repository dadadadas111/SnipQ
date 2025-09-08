export namespace main {
	
	export class AppSettings {
	    triggerPrefix: string;
	    expandKey: string;
	    typingTimeout: number;
	    expansionMethod: string;
	    suggestionsEnabled: boolean;
	    minQueryLength: number;
	    maxSuggestions: number;
	    suggestionDelay: number;
	    suggestionHideDelay: number;
	    strictBoundaries: boolean;
	    pauseOnFailure: boolean;
	    bufferSize: number;
	    caseSensitive: boolean;
	    hotkeys?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.triggerPrefix = source["triggerPrefix"];
	        this.expandKey = source["expandKey"];
	        this.typingTimeout = source["typingTimeout"];
	        this.expansionMethod = source["expansionMethod"];
	        this.suggestionsEnabled = source["suggestionsEnabled"];
	        this.minQueryLength = source["minQueryLength"];
	        this.maxSuggestions = source["maxSuggestions"];
	        this.suggestionDelay = source["suggestionDelay"];
	        this.suggestionHideDelay = source["suggestionHideDelay"];
	        this.strictBoundaries = source["strictBoundaries"];
	        this.pauseOnFailure = source["pauseOnFailure"];
	        this.bufferSize = source["bufferSize"];
	        this.caseSensitive = source["caseSensitive"];
	        this.hotkeys = source["hotkeys"];
	    }
	}

}

export namespace types {
	
	export class Group {
	    id: string;
	    name: string;
	    description?: string;
	    icon?: string;
	    order?: number;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.order = source["order"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Hotkey {
	    modifiers: string[];
	    key: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Hotkey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.modifiers = source["modifiers"];
	        this.key = source["key"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Rendered {
	    output: string;
	    cursorOffset: number;
	    usedSnippet: string;
	    usedParams: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Rendered(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output = source["output"];
	        this.cursorOffset = source["cursorOffset"];
	        this.usedSnippet = source["usedSnippet"];
	        this.usedParams = source["usedParams"];
	    }
	}
	export class Settings {
	    prefix: string;
	    expandKey: string;
	    strictBoundaries: boolean;
	    excludedApps?: string[];
	    locale: string;
	    defaultDateFormat: string;
	    timezone: string;
	    historyEnabled: boolean;
	    historyLimit: number;
	    pinForSensitive: boolean;
	    hotkeys?: Record<string, Hotkey>;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prefix = source["prefix"];
	        this.expandKey = source["expandKey"];
	        this.strictBoundaries = source["strictBoundaries"];
	        this.excludedApps = source["excludedApps"];
	        this.locale = source["locale"];
	        this.defaultDateFormat = source["defaultDateFormat"];
	        this.timezone = source["timezone"];
	        this.historyEnabled = source["historyEnabled"];
	        this.historyLimit = source["historyLimit"];
	        this.pinForSensitive = source["pinForSensitive"];
	        this.hotkeys = this.convertValues(source["hotkeys"], Hotkey, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Snippet {
	    id: string;
	    name: string;
	    trigger: string;
	    description?: string;
	    tags?: string[];
	    strict?: boolean;
	    defaults?: Record<string, any>;
	    template: string;
	    groupId: string;
	
	    static createFrom(source: any = {}) {
	        return new Snippet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.trigger = source["trigger"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.strict = source["strict"];
	        this.defaults = source["defaults"];
	        this.template = source["template"];
	        this.groupId = source["groupId"];
	    }
	}

}

