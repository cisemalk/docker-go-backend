TÜRKÇE
Kodu çalıştırmak için indirilmesi gerekenler:
Docker ve Git Bash
Bunlar indirildikten sonra Docker ve Git Bash açılmalı. Directory olarak backend isimli klasörün olduğu yer yazılmalı. (Örnek: $ cd Desktop $ cd backend) Bu işlem gerçekleştiğinde ilk olarak Git Bash'e:
docker-compose build komutu yazılmalıdır. İşlem bittikten sonra Docker'ın açılması için:
docker-compose up -d komutu yazılmalıdır. Eğer Docker kontrol edildiğinde eksik parçalar gözükürse (Ekranda 2/3 yazması gibi) 
docker-compose stop app && docker-compose rm app -f && docker-compose up -d app komutu kullanılarak sorun halledilir.

Bu işlemler bittiğinde Android Studio'ya yeni bir Flutter projesi oluşturularak bizim tarafımızdan yapılan Flutter dosyasını koyarak uygulamanın arayüzü bir emülatör aracılığıyla kullanılabilir. 
NOT: Uygulamayı test ederseniz plane ve ticket databaselerin boş olduğunu fark edersiniz. Onlara içerik eklemek için localhosta gidip:

INSERT INTO `planes` (`id`, `created_at`, `updated_at`, `deleted_at`, `firm_name`, `seat_number`) VALUES (NULL, NULL, NULL, NULL, 'A', '2'), (NULL, NULL, NULL, NULL, 'B', '4');

INSERT INTO `tickets` (`id`, `created_at`, `updated_at`, `deleted_at`, `plane_id`, `from`, `to`, `departure_date`, `return_date`, `d_hour`, `r_hour`, `nof_seats`, `price`) VALUES (NULL, NULL, NULL, NULL, '2', 'Ankara', 'İstanbul', '2023-08-08', '2023-08-09', '19:00', '22:00', '66', '111'), (NULL, NULL, NULL, NULL, '1', 'Ankara', 'İstanbul', '2023-08-08', '2023-08-09', '16:00', '18:30', '100', '665');

örnek sql kodlarını kullanabilirsiniz.

ENGLISH:
You have to download these in order to execute the code:
Docker and Git Bash
Once these are downloaded, Docker and Git Bash should open. As the directory, the location of the folder named backend should be written.(For example: $ cd Desktop $ cd backend) When this happens, first go to Git Bash:
Docker-compose build command should be written. To open Docker after the process is finished:
Docker-compose up -d command should be typed. If the Docker check shows missing parts (like 2/3 on the screen)
Using the command docker-compose stop app && docker-compose rm app -f && docker-compose up -d app will fix the problem.

When these processes are finished, a new Flutter project is created in Android Studio and the interface of the application can be used via an emulator by putting the Flutter file made by us.

NOTE: If you test the application, you will notice that the plane and ticket databases are empty. You can go to localhost and use:

INSERT INTO `planes` (`id`, `created_at`, `updated_at`, `deleted_at`, `firm_name`, `seat_number`) VALUES (NULL, NULL, NULL, NULL, 'A', '2'), (NULL, NULL, NULL, NULL, 'B', '4');

INSERT INTO `tickets` (`id`, `created_at`, `updated_at`, `deleted_at`, `plane_id`, `from`, `to`, `departure_date`, `return_date`, `d_hour`, `r_hour`, `nof_seats`, `price`) VALUES (NULL, NULL, NULL, NULL, '2', 'Ankara', 'İstanbul', '2023-08-08', '2023-08-09', '19:00', '22:00', '66', '111'), (NULL, NULL, NULL, NULL, '1', 'Ankara', 'İstanbul', '2023-08-08', '2023-08-09', '16:00', '18:30', '100', '665');

these sample sql codes to add content.



