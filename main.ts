const fs = require('fs');
const JavaToTypescriptConverter = require('java2typescript');

function main(){

    let folderRegEx = new RegExp("^dombler-fx-core");

    fs.readdir("..", function (err, files){
            files.forEach(async (file) => {
                // will also include directory names
                if (folderNameStartWithPattern(file, folderRegEx)) {
                    const configuration = {
                        packageRoot: "../" + file,
                    };

                    const converter = new JavaToTypescriptConverter(configuration);
                    await converter.startConversion();
                }
            });
        });
}

function folderNameStartWithPattern(folder, pattern){
    if (pattern.test(folder))
        return true;
    else
        return false;
}