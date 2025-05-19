export namespace types {
	
	export class CrawlingRequest {
	    img_save_path: string;
	    img_urls: string[];
	    timeout_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new CrawlingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.img_save_path = source["img_save_path"];
	        this.img_urls = source["img_urls"];
	        this.timeout_seconds = source["timeout_seconds"];
	    }
	}
	export class CrawlingResponse {
	    text_content_save_path: string;
	    crawl_url_count: number;
	    crawl_img_count: number;
	    err_content: string;
	    cast_time_str: string;
	
	    static createFrom(source: any = {}) {
	        return new CrawlingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text_content_save_path = source["text_content_save_path"];
	        this.crawl_url_count = source["crawl_url_count"];
	        this.crawl_img_count = source["crawl_img_count"];
	        this.err_content = source["err_content"];
	        this.cast_time_str = source["cast_time_str"];
	    }
	}
	export class CroppingRequest {
	    img_save_path: string;
	    bottom_pixel: number;
	
	    static createFrom(source: any = {}) {
	        return new CroppingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.img_save_path = source["img_save_path"];
	        this.bottom_pixel = source["bottom_pixel"];
	    }
	}
	export class CroppingResponse {
	    crop_img_path: string;
	    crop_img_count: number;
	    err_content: string;
	    cast_time_str: string;
	
	    static createFrom(source: any = {}) {
	        return new CroppingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.crop_img_path = source["crop_img_path"];
	        this.crop_img_count = source["crop_img_count"];
	        this.err_content = source["err_content"];
	        this.cast_time_str = source["cast_time_str"];
	    }
	}
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
	export class ShufflingRequest {
	    img_save_path: string;
	    max_num_image: number;
	
	    static createFrom(source: any = {}) {
	        return new ShufflingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.img_save_path = source["img_save_path"];
	        this.max_num_image = source["max_num_image"];
	    }
	}
	export class ShufflingResponse {
	    shuffle_img_path: string;
	    cast_time_str: string;
	
	    static createFrom(source: any = {}) {
	        return new ShufflingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shuffle_img_path = source["shuffle_img_path"];
	        this.cast_time_str = source["cast_time_str"];
	    }
	}

}

