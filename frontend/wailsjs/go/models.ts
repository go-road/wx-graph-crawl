export namespace types {
	
	export class PreferenceSet {
	    save_img_path: string;
	    download_timeout: number;
	    crop_img_bottom_pixel: number;
	
	    static createFrom(source: any = {}) {
	        return new PreferenceSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.save_img_path = source["save_img_path"];
	        this.download_timeout = source["download_timeout"];
	        this.crop_img_bottom_pixel = source["crop_img_bottom_pixel"];
	    }
	}
	export class SelectFileResponse {
	    file_path: string;
	    valid_urls: string[];
	
	    static createFrom(source: any = {}) {
	        return new SelectFileResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_path = source["file_path"];
	        this.valid_urls = source["valid_urls"];
	    }
	}

}

