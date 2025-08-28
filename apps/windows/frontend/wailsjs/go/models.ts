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

